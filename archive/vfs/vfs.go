package vfs

import (
	"fmt"
	"path"
	"strings"

	"github.com/Merith-TK/resonite-sh/pkg/objects"
	"github.com/Merith-TK/resonite-sh/pkg/resolink"
)

// VFS represents the virtual filesystem
type VFS struct {
	client      *resolink.Client
	root        *Node
	currentPath string
	cache       map[string]*Node
}

// Node represents a node in the virtual filesystem
type Node struct {
	Slot     *objects.Slot
	Parent   *Node
	Children map[string]*Node
	Loaded   bool
}

// NewVFS creates a new virtual filesystem
func NewVFS(client *resolink.Client) *VFS {
	return &VFS{
		client:      client,
		currentPath: "/",
		cache:       make(map[string]*Node),
	}
}

// Initialize loads the root slot
func (vfs *VFS) Initialize() error {
	// TODO: Get actual root slot ID from Resonite
	// For now, we'll use a placeholder
	rootSlot := objects.NewSlot("root", "Root", "")

	vfs.root = &Node{
		Slot:     rootSlot,
		Parent:   nil,
		Children: make(map[string]*Node),
		Loaded:   false,
	}

	vfs.cache["/"] = vfs.root
	return nil
}

// ResolvePath resolves a path to a node
func (vfs *VFS) ResolvePath(p string) (*Node, error) {
	// Handle absolute vs relative paths
	if !strings.HasPrefix(p, "/") {
		// Relative path - resolve from current directory
		p = path.Join(vfs.currentPath, p)
	}

	// Clean the path
	p = path.Clean(p)

	// Check cache
	if node, exists := vfs.cache[p]; exists {
		return node, nil
	}

	// Navigate from root
	parts := strings.Split(strings.Trim(p, "/"), "/")
	current := vfs.root

	for i, part := range parts {
		if part == "" {
			continue
		}

		// Ensure current node is loaded
		if !current.Loaded {
			if err := vfs.loadNode(current); err != nil {
				return nil, fmt.Errorf("failed to load node: %w", err)
			}
		}

		// Find child
		child, exists := current.Children[part]
		if !exists {
			return nil, fmt.Errorf("path not found: %s", p)
		}

		current = child

		// Cache intermediate paths
		intermediatePath := "/" + strings.Join(parts[:i+1], "/")
		vfs.cache[intermediatePath] = current
	}

	return current, nil
}

// loadNode loads children and components for a node
func (vfs *VFS) loadNode(node *Node) error {
	// Get slot details with components
	resp, err := vfs.client.GetSlot(node.Slot.ID, true)
	if err != nil {
		return err
	}

	// TODO: Parse response and populate node.Slot.Components

	// List children
	childrenResp, err := vfs.client.ListChildren(node.Slot.ID)
	if err != nil {
		return err
	}

	// TODO: Parse children response and create child nodes
	_ = childrenResp

	node.Loaded = true
	return nil
}

// GetCurrentPath returns the current working directory
func (vfs *VFS) GetCurrentPath() string {
	return vfs.currentPath
}

// ChangeDirectory changes the current directory
func (vfs *VFS) ChangeDirectory(p string) error {
	node, err := vfs.ResolvePath(p)
	if err != nil {
		return err
	}

	// Build new path
	newPath := vfs.buildPath(node)
	vfs.currentPath = newPath

	return nil
}

// buildPath constructs the full path for a node
func (vfs *VFS) buildPath(node *Node) string {
	if node == vfs.root {
		return "/"
	}

	parts := []string{}
	current := node

	for current != nil && current != vfs.root {
		parts = append([]string{current.Slot.Name}, parts...)
		current = current.Parent
	}

	return "/" + strings.Join(parts, "/")
}

// ListDirectory lists contents of a directory
func (vfs *VFS) ListDirectory(p string) ([]*Node, error) {
	node, err := vfs.ResolvePath(p)
	if err != nil {
		return nil, err
	}

	// Ensure node is loaded
	if !node.Loaded {
		if err := vfs.loadNode(node); err != nil {
			return nil, err
		}
	}

	// Collect children
	children := make([]*Node, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child)
	}

	return children, nil
}

// InvalidateCache clears cached node data
func (vfs *VFS) InvalidateCache(p string) {
	delete(vfs.cache, p)
}
