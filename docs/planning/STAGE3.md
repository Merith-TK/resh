# Stage 3 Implementation: Component Inspection

## Overview
Stage 3 adds the ability to inspect and edit component members in the Resonite Shell.

## Architecture

### Business Logic (pkg/shell)
- **component_models.go**: Data structures for component inspection
  - `ComponentData`: Full component with parsed type name and members
  - `MemberData`: Individual component member with name, ID, type, and value

- **component_operations.go**: Component-related operations
  - `InspectComponent()`: Fetches component data and parses members
  - `SetComponentMember()`: Updates a member value (placeholder for now)
  - `parseComponentTypeName()`: Strips namespace prefix from type names

### Display Layer (cmd)
- **display.go**: Added component display functions
  - `displayComponentData()`: Renders full component inspection with formatted members
  - `displayMemberInfo()`: Renders individual member with alignment
  - `formatMemberValue()`: Formats member values for display (null, bool, number, string)

- **commands.go**: Added component commands
  - `inspectComponent()`: Handler for inspect command
  - `setComponentMember()`: Handler for set command (placeholder)

- **repl.go**: Added command handlers
  - `inspect <id>`: Inspect a component's members
  - `set <id> <value>`: Set a member value

## Usage Example

```
$ ./resonite-sh repl
Connected to ResoLink at ws://localhost:39015

Root $ ls
P [slot] ID_73 __TEMP
P [comp] ID_9 StaticLocaleProvider

Root $ inspect ID_9

Component: StaticLocaleProvider ID_9

ID_242 [bool  ] persistent    = true
ID_243 [int   ] UpdateOrder    = 0
ID_244 [bool  ] Enabled        = true
ID_245 [Uri   ] URL            = "builtin-locale://core/"
ID_246 [string] OverrideLocale = <null>

Root $ set ID_9 ID_244=false
✓ Member updated

Root $ inspect ID_9

Component: StaticLocaleProvider ID_9

ID_242 [bool  ] persistent    = true
ID_243 [int   ] UpdateOrder    = 0
ID_244 [bool  ] Enabled        = false
ID_245 [Uri   ] URL            = "builtin-locale://core/"
ID_246 [string] OverrideLocale = <null>
```

## Component Member Structure

Components in Resonite have this structure:
```json
{
  "id": "Reso_9",
  "componentType": "[FrooxEngine]FrooxEngine.StaticLocaleProvider",
  "members": {
    "Enabled": {
      "$type": "bool",
      "id": "Reso_244",
      "value": true
    },
    "URL": {
      "$type": "Uri",
      "id": "Reso_245",
      "value": "builtin-locale://core/"
    }
  }
}
```

## Implementation Status

### ✅ Completed
- Component data structures (ComponentData, MemberData)
- InspectComponent operation to fetch and parse components
- SetComponentMember operation with type conversion
- Display formatting with alignment and color coding
- Inspect command in REPL
- Set command with `component_id member_id=value` syntax
- Type name parsing (strips [FrooxEngine]FrooxEngine. prefix)
- Value formatting for common types (bool, int, string, null)
- Type conversion for set command (bool, int, float, string, Uri)

### ⬜ Planned
- Support for complex types (float3, floatQ, color, etc.)
- Batch editing (multiple member_id=value pairs in one command)
- TUI editor for visual component editing
- Bulk editing operations
- Undo/redo for component changes

## Notes

- **ID Format**: Display shows `ID_xxx` but internally uses `Reso_xxx`
- **Type Parsing**: Component types like `[FrooxEngine]FrooxEngine.StaticLocaleProvider` are shown as `StaticLocaleProvider`
- **Member Types**: Common types include `bool`, `int`, `string`, `Uri`, `float`, `float3`, `floatQ`, `color`
- **Null Values**: Displayed as `<null>` for clarity
- **Set Command Syntax**: `set <component_id> <member_id>=<value>` - Both component and member must be specified
- **Type Conversion**: Values are automatically converted based on member type (bool, int, float, string)
- **Architecture**: CLI commands test operations first, TUI will reuse pkg/shell operations later
