# ResoniteLink Protocol Reference

This document provides a reference for the ResoniteLink WebSocket protocol based on the official C# implementation.

## Connection

**WebSocket URL**: `ws://localhost:29551` (default)

All messages are JSON formatted with a `$type` field indicating the message type.

## Message Structure

### Request Format
```json
{
    "$type": "messageType",
    "messageId": "unique-id",
    ...additional fields...
}
```

### Response Format
```json
{
    "$type": "response",
    "messageId": "matching-request-id",
    ...response data...
}
```

## Slot Operations

### Get Slot
**Message Type**: `getSlot`

```json
{
    "$type": "getSlot",
    "messageId": "msg-1",
    "slotId": "Root",
    "includeComponentData": false,
    "depth": 0
}
```

**Fields**:
- `slotId`: Slot ID (use `"Root"` for root slot)
- `includeComponentData`: Include full component data (default: false)
- `depth`: How deep to fetch (-1 = all children, 0 = only this slot)

**Response**: `SlotData` with slot information

### Add Slot
**Message Type**: `addSlot`

```json
{
    "$type": "addSlot",
    "messageId": "msg-2",
    "data": {
        "id": "MyCustomID",
        "parent": {
            "$type": "reference",
            "targetId": "Root"
        },
        "name": {
            "$type": "string",
            "value": "MySlot"
        },
        "position": {
            "$type": "float3",
            "value": {"x": 0, "y": 1.5, "z": 0}
        },
        "active": {
            "$type": "bool",
            "value": true
        },
        "persistent": {
            "$type": "bool",
            "value": true
        }
    }
}
```

**Notes**:
- `id` is optional; Resonite will auto-generate if omitted
- Only include fields you want to set; others use defaults
- Parent defaults to Root if omitted

### Update Slot
**Message Type**: `updateSlot`

```json
{
    "$type": "updateSlot",
    "messageId": "msg-3",
    "data": {
        "id": "MyCustomID",
        "name": {
            "$type": "string",
            "value": "NewName"
        },
        "scale": {
            "$type": "float3",
            "value": {"x": 2, "y": 2, "z": 2}
        }
    }
}
```

**Notes**:
- `id` is MANDATORY for updates
- Only include fields you want to change

### Remove Slot
**Message Type**: `removeSlot`

```json
{
    "$type": "removeSlot",
    "messageId": "msg-4",
    "slotId": "MyCustomID"
}
```

## Component Operations

### Get Component
**Message Type**: `getComponent`

```json
{
    "$type": "getComponent",
    "messageId": "msg-5",
    "componentId": "IDabc123"
}
```

### Add Component
**Message Type**: `addComponent`

```json
{
    "$type": "addComponent",
    "messageId": "msg-6",
    "containerSlotId": "MyCustomID",
    "data": {
        "id": "MyComponentID",
        "componentType": "[FrooxEngine]FrooxEngine.BoxMesh",
        "members": {
            "Persistent": {
                "$type": "bool",
                "value": true
            },
            "Size": {
                "$type": "float3",
                "value": {"x": 1, "y": 1, "z": 1}
            }
        }
    }
}
```

**Notes**:
- `containerSlotId` is MANDATORY
- `componentType` is MANDATORY
- `id` is optional (auto-generated if omitted)
- Component type format: `[AssemblyName]Namespace.ClassName`

### Update Component
**Message Type**: `updateComponent`

```json
{
    "$type": "updateComponent",
    "messageId": "msg-7",
    "data": {
        "id": "MyComponentID",
        "members": {
            "Size": {
                "$type": "float3",
                "value": {"x": 2, "y": 2, "z": 2}
            }
        }
    }
}
```

**Notes**:
- `id` is MANDATORY for updates
- Only include members you want to change

### Remove Component
**Message Type**: `removeComponent`

```json
{
    "$type": "removeComponent",
    "messageId": "msg-8",
    "componentId": "MyComponentID"
}
```

## Reflection / Type Discovery

### Get Component Type List
**Message Type**: `getComponentTypeList`

```json
{
    "$type": "getComponentTypeList",
    "messageId": "msg-9"
}
```

Returns list of all available component types.

### Get Component Definition
**Message Type**: `getComponentDefinition`

```json
{
    "$type": "getComponentDefinition",
    "messageId": "msg-10",
    "componentType": "[FrooxEngine]FrooxEngine.BoxMesh"
}
```

Returns detailed information about component including all members, types, and defaults.

### Get Type Definition
**Message Type**: `getTypeDefinition`

```json
{
    "$type": "getTypeDefinition",
    "messageId": "msg-11",
    "typeName": "float3"
}
```

### Get Enum Definition
**Message Type**: `getEnumDefinition`

```json
{
    "$type": "getEnumDefinition",
    "messageId": "msg-12",
    "enumType": "FrooxEngine.ColorProfile"
}
```

## Session Data

### Request Session Data
**Message Type**: `requestSessionData`

```json
{
    "$type": "requestSessionData",
    "messageId": "msg-13"
}
```

Returns information about the current session.

## Batch Operations

### Data Model Operation Batch
**Message Type**: `dataModelOperationBatch`

```json
{
    "$type": "dataModelOperationBatch",
    "messageId": "msg-14",
    "operations": [
        {
            "$type": "addSlot",
            "data": {...}
        },
        {
            "$type": "addComponent",
            "containerSlotId": "...",
            "data": {...}
        }
    ]
}
```

Allows executing multiple operations in a single message for better performance.

## Data Types

### Primitive Types

#### bool
```json
{
    "$type": "bool",
    "value": true
}
```

#### int
```json
{
    "$type": "int",
    "value": 42
}
```

#### float
```json
{
    "$type": "float",
    "value": 3.14159
}
```

#### string
```json
{
    "$type": "string",
    "value": "Hello World"
}
```

#### long
```json
{
    "$type": "long",
    "value": 999
}
```

### Vector Types

#### float2
```json
{
    "$type": "float2",
    "value": {"x": 1.0, "y": 2.0}
}
```

#### float3 (Position, Scale, etc.)
```json
{
    "$type": "float3",
    "value": {"x": 0, "y": 1.5, "z": 0}
}
```

#### float4
```json
{
    "$type": "float4",
    "value": {"x": 1, "y": 0, "z": 0, "w": 1}
}
```

#### floatQ (Quaternion/Rotation)
```json
{
    "$type": "floatQ",
    "value": {"x": 0, "y": 0, "z": 0, "w": 1}
}
```

### Color Types

#### color
```json
{
    "$type": "color",
    "value": {"r": 1, "g": 0, "b": 0, "a": 1}
}
```

#### colorX
```json
{
    "$type": "colorX",
    "value": {"r": 1, "g": 0, "b": 0, "a": 1}
}
```

### Reference Types

#### reference (Slot or Component reference)
```json
{
    "$type": "reference",
    "targetId": "ID2300"
}
```

### List Types

#### list
```json
{
    "$type": "list",
    "elements": [
        {
            "$type": "reference",
            "targetId": "MaterialID1"
        },
        {
            "$type": "reference",
            "targetId": "MaterialID2"
        }
    ]
}
```

**Note**: When adding list elements, omit the element `id`. To update elements, you must include both the element `id` and data (two-stage update pattern).

## Common Component Types

### Visual Components
- `[FrooxEngine]FrooxEngine.BoxMesh` - Box mesh
- `[FrooxEngine]FrooxEngine.SphereMesh` - Sphere mesh
- `[FrooxEngine]FrooxEngine.MeshRenderer` - Renders meshes
- `[FrooxEngine]FrooxEngine.PBS_Metallic` - PBR material

### Transform & Physics
- `[FrooxEngine]FrooxEngine.Grabbable` - Make object grabbable
- `[FrooxEngine]FrooxEngine.BoxCollider` - Box collision

### UIX Components
- `[FrooxEngine]FrooxEngine.UIX.Canvas` - UI canvas
- `[FrooxEngine]FrooxEngine.UIX.RectTransform` - UI layout
- `[FrooxEngine]FrooxEngine.UIX.Text` - UI text
- `[FrooxEngine]FrooxEngine.UIX.Image` - UI image
- `[FrooxEngine]FrooxEngine.UIX.Button` - UI button

### Dynamic Variables
- `[FrooxEngine]FrooxEngine.DynamicValueVariable<T>` - Value variable
- `[FrooxEngine]FrooxEngine.DynamicObjectVariable<T>` - Object variable
- `[FrooxEngine]FrooxEngine.DynamicReferenceVariable<T>` - Reference variable
- `[FrooxEngine]FrooxEngine.DynamicVariableSpace` - Variable namespace

### ProtoFlux
- `[ProtoFluxBindings]FrooxEngine.ProtoFlux.Runtimes.Execution.Nodes.*` - ProtoFlux nodes
- Generic types use `<>` notation: `ValueInput<int>`, `ValueAdd<float>`

## Slot Properties

Standard slot properties available in slot data:

- `id` - Slot RefID
- `name` - Slot name (string)
- `active` - Is slot active (bool)
- `persistent` - Is slot persistent (bool)
- `position` - Local position (float3)
- `rotation` - Local rotation (floatQ)
- `scale` - Local scale (float3)
- `orderOffset` - Rendering order offset (long)
- `tag` - Slot tag (string)
- `parent` - Parent slot reference
- `components` - List of component references
- `children` - List of child slot references

## Error Handling

Responses may include error information:

```json
{
    "$type": "error",
    "messageId": "msg-15",
    "error": "Error message here"
}
```

Common errors:
- Invalid slot/component ID
- Invalid component type
- Invalid field name or value
- Permission denied
- Slot/component already removed

## Best Practices

1. **Use messageId**: Always provide unique messageId to correlate responses
2. **Batch operations**: Use `dataModelOperationBatch` for multiple operations
3. **Lazy loading**: Use `depth: 0` and load children on-demand
4. **Cache data**: Cache slot/component data locally to reduce requests
5. **Two-stage list updates**: When updating list elements, first add without `id`, then update with `id`
6. **Component types**: Use `getComponentTypeList` to discover available types
7. **Validation**: Use reflection APIs to validate member names and types

## RESH Slot Creation Example

Complete example of creating the RESH variable storage slot:

```json
{
    "$type": "addSlot",
    "messageId": "create-resh",
    "data": {
        "id": "RESH_SLOT",
        "parent": {"$type": "reference", "targetId": "Root"},
        "name": {"$type": "string", "value": "RESH"},
        "persistent": {"$type": "bool", "value": true},
        "position": {"$type": "float3", "value": {"x": 0, "y": 0, "z": 0}},
        "rotation": {"$type": "floatQ", "value": {"x": 0, "y": 0, "z": 0, "w": 1}},
        "scale": {"$type": "float3", "value": {"x": 1, "y": 1, "z": 1}},
        "orderOffset": {"$type": "long", "value": 999}
    }
}
```

Then add DynamicVariableSpace:

```json
{
    "$type": "addComponent",
    "messageId": "add-dynvarspace",
    "containerSlotId": "RESH_SLOT",
    "data": {
        "componentType": "[FrooxEngine]FrooxEngine.DynamicVariableSpace",
        "members": {
            "Persistent": {"$type": "bool", "value": true},
            "SpaceName": {"$type": "string", "value": "RESH"},
            "OnlyDirectBinding": {"$type": "bool", "value": true}
        }
    }
}
```

Then add DynamicReferenceVariable:

```json
{
    "$type": "addComponent",
    "messageId": "add-refvar",
    "containerSlotId": "RESH_SLOT",
    "data": {
        "componentType": "[FrooxEngine]FrooxEngine.DynamicReferenceVariable<Slot>",
        "members": {
            "Persistent": {"$type": "bool", "value": true},
            "VariableName": {"$type": "string", "value": "World/RESH.DATA"},
            "Reference": {"$type": "reference", "targetId": "RESH_SLOT"}
        }
    }
}
```

## Notes

- All operations require the session to be in "Host" mode
- Some operations may be restricted based on world permissions
- RefIDs use format: `ID` + alphanumeric (e.g., `ID2300`, `IDabc123`)
- Root slot can be referenced as `"Root"` string or typically `"ID2300"`
