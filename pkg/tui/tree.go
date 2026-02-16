package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Merith-TK/resh/pkg/shell"
)

// LoadTreeItems loads the children and components of the focused slot
func (m *Model) LoadTreeItems() error {
	// Get slot listing
	listing, err := shell.ListSlot(m.Client, m.FocusedSlotID, m.State.RootSlotID)
	if err != nil {
		return fmt.Errorf("failed to load tree items: %w", err)
	}

	m.TreeItems = []TreeItem{}

	// Filter out RESH.DATA if configured
	children := listing.Children
	if m.HideRESHData {
		filtered := []shell.SlotInfo{}
		for _, child := range children {
			if child.Name != "RESH.DATA" && child.Name != "RESH" {
				filtered = append(filtered, child)
			}
		}
		children = filtered
	}

	// Sort children by name
	sort.Slice(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})

	// Add child slots first
	for _, child := range children {
		m.TreeItems = append(m.TreeItems, TreeItem{
			ID:       child.ID,
			Name:     child.Name,
			IsSlot:   true,
			IsActive: child.IsActive,
		})
	}

	// Add components if not hiding root components
	if !m.HideRoot || m.FocusedSlotID != m.State.RootSlotID {
		// Sort components by type
		components := listing.Components
		sort.Slice(components, func(i, j int) bool {
			return components[i].Type < components[j].Type
		})

		for _, comp := range components {
			// Shorten component type name
			typeName := comp.Type
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				typeName = parts[len(parts)-1]
			}

			m.TreeItems = append(m.TreeItems, TreeItem{
				ID:     comp.ID,
				Name:   typeName,
				IsSlot: false,
				Type:   comp.Type,
			})
		}
	}

	// Reset cursor if out of bounds
	if m.TreeCursor >= len(m.TreeItems) {
		m.TreeCursor = 0
	}

	return nil
}

// RenderTree renders the tree panel
func (m Model) RenderTree(width, height int) string {
	var b strings.Builder

	// Title with breadcrumb path
	pathStr := strings.Join(m.PathBreadcrumb, " > ")
	if len(pathStr) > width-10 {
		// Truncate from the left, keep the end
		pathStr = "..." + pathStr[len(pathStr)-(width-13):]
	}
	title := titleStyle.Render(fmt.Sprintf("Tree: %s", pathStr))
	b.WriteString(title)
	b.WriteString("\n")

	// Calculate available height for content (subtract title and padding)
	contentHeight := height - 4

	// Render tree items
	if len(m.TreeItems) == 0 {
		b.WriteString(helpStyle.Render("  (empty)"))
	} else {
		// Calculate scroll offset to keep cursor visible
		scrollOffset := 0
		if m.TreeCursor >= contentHeight {
			scrollOffset = m.TreeCursor - contentHeight + 1
		}

		// Render visible items
		endIdx := scrollOffset + contentHeight
		if endIdx > len(m.TreeItems) {
			endIdx = len(m.TreeItems)
		}

		for i := scrollOffset; i < endIdx; i++ {
			item := m.TreeItems[i]
			line := m.renderTreeItem(item, i == m.TreeCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Help text at bottom
	b.WriteString("\n")
	if m.Focus == FocusTree {
		b.WriteString(helpStyle.Render("  ↑/↓:nav Enter:select →/f:focus ←/b:parent r:reload Tab:inspector"))
	}

	content := b.String()
	return ApplyBorder(content, m.Focus == FocusTree, width-2, height-2)
}

// renderTreeItem renders a single tree item
func (m Model) renderTreeItem(item TreeItem, selected bool) string {
	var icon string

	if item.IsSlot {
		icon = "├─"
		if !item.IsActive {
			icon = "├─[X]" // Indicate inactive
		}
	} else {
		icon = "├─[C]"
	}

	// Build the line
	name := item.Name
	if len(name) > 35 {
		name = name[:32] + "..."
	}

	line := fmt.Sprintf("  %s %s", icon, name)

	// Apply selection style
	if selected {
		return selectedTreeItem.Render(line)
	}
	return normalTreeItem.Render(line)
}

// MoveCursorUp moves the tree cursor up
func (m *Model) MoveCursorUp() {
	if m.TreeCursor > 0 {
		m.TreeCursor--
		if m.UpdateOnNavigate {
			m.SelectCurrentItem()
		}
	}
}

// MoveCursorDown moves the tree cursor down
func (m *Model) MoveCursorDown() {
	if m.TreeCursor < len(m.TreeItems)-1 {
		m.TreeCursor++
		if m.UpdateOnNavigate {
			m.SelectCurrentItem()
		}
	}
}

// SelectCurrentItem loads the current tree item into inspector
func (m *Model) SelectCurrentItem() {
	if m.TreeCursor >= 0 && m.TreeCursor < len(m.TreeItems) {
		item := m.TreeItems[m.TreeCursor]
		m.InspectedItem = &item
		m.FieldCursor = 0
		m.EditingField = false

		// Load data using shell.InspectItem
		data, isSlot, err := shell.InspectItem(m.Client, item.ID)
		if err != nil {
			m.ErrorMessage = fmt.Sprintf("Failed to inspect: %v", err)
			return
		}

		// Verify item type matches
		if isSlot != item.IsSlot {
			m.ErrorMessage = "Item type mismatch"
			return
		}

		m.InspectedData = data
		m.ErrorMessage = ""
	}
}

// FocusOnSlot changes the tree focus to the selected slot
func (m *Model) FocusOnSlot() error {
	if m.TreeCursor >= 0 && m.TreeCursor < len(m.TreeItems) {
		item := m.TreeItems[m.TreeCursor]
		if !item.IsSlot {
			m.ErrorMessage = "Can only focus on slots (not components)"
			return fmt.Errorf("not a slot")
		}

		m.FocusedSlotID = item.ID
		m.FocusedSlotName = item.Name
		m.TreeCursor = 0
		m.ErrorMessage = "" // Clear any previous errors

		// Update breadcrumb path
		m.PathBreadcrumb = append(m.PathBreadcrumb, item.Name)

		// Reload tree
		if err := m.LoadTreeItems(); err != nil {
			m.ErrorMessage = fmt.Sprintf("Failed to load slot: %v", err)
			// Revert breadcrumb on error
			m.PathBreadcrumb = m.PathBreadcrumb[:len(m.PathBreadcrumb)-1]
			return err
		}

		m.StatusMessage = fmt.Sprintf("Focused on: %s", item.Name)
	}
	return nil
}

// FocusOnParent changes the tree focus to the parent slot
func (m *Model) FocusOnParent() error {
	// Can't go above root
	if m.FocusedSlotID == m.State.RootSlotID || m.FocusedSlotID == "Root" {
		m.ErrorMessage = "Already at root (can't go higher)"
		return fmt.Errorf("already at root")
	}

	// Get current slot to find parent
	slotResp, err := m.Client.GetSlot(m.FocusedSlotID, false, 0)
	if err != nil {
		m.ErrorMessage = fmt.Sprintf("Failed to get slot: %v", err)
		return fmt.Errorf("failed to get slot: %w", err)
	}

	if slotResp.Data.Parent == nil {
		m.ErrorMessage = "No parent slot found"
		return fmt.Errorf("no parent")
	}

	parentID := slotResp.Data.Parent.TargetID

	// Get parent name
	parentResp, err := m.Client.GetSlot(parentID, false, 0)
	if err != nil {
		m.ErrorMessage = fmt.Sprintf("Failed to get parent: %v", err)
		return fmt.Errorf("failed to get parent: %w", err)
	}

	parentName := "Root"
	if parentResp.Data.Name != nil {
		parentName = parentResp.Data.Name.Value
	}

	m.FocusedSlotID = parentID
	m.FocusedSlotName = parentName
	m.TreeCursor = 0
	m.ErrorMessage = "" // Clear errors

	// Update breadcrumb: pop last item
	if len(m.PathBreadcrumb) > 1 {
		m.PathBreadcrumb = m.PathBreadcrumb[:len(m.PathBreadcrumb)-1]
	}

	// Reload tree
	if err := m.LoadTreeItems(); err != nil {
		m.ErrorMessage = fmt.Sprintf("Failed to load parent: %v", err)
		return err
	}

	m.StatusMessage = fmt.Sprintf("Parent: %s", parentName)
	return nil
}
