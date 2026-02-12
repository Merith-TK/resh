package resolink

import (
	"encoding/json"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		msg     interface{}
		wantErr bool
	}{
		{
			name: "GetSlot message",
			msg: &GetSlotMessage{
				BaseMessage: BaseMessage{
					MessageID: "test-1",
					Type:      "getSlot",
				},
				SlotID:               "Root",
				IncludeComponentData: false,
				Depth:                0,
			},
			wantErr: false,
		},
		{
			name: "AddSlot message with values",
			msg: &AddSlotMessage{
				BaseMessage: BaseMessage{
					MessageID: "test-2",
					Type:      "addSlot",
				},
				Data: &SlotDefinition{
					ID:       "TestSlot1",
					Name:     NewValueString("My Test Slot"),
					Position: NewValueFloat3(0, 1.5, 0),
					Active:   NewValueBool(true),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				t.Logf("Serialized: %s", string(data))

				// Verify it has $type field
				var check map[string]interface{}
				if err := json.Unmarshal(data, &check); err != nil {
					t.Errorf("Failed to unmarshal back: %v", err)
					return
				}

				if _, ok := check["$type"]; !ok {
					t.Error("Missing $type field in serialized message")
				}
			}
		})
	}
}

func TestValueTypeCreation(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", NewValueString("test"), "string"},
		{"bool", NewValueBool(true), "bool"},
		{"int", NewValueInt(42), "int"},
		{"long", NewValueLong(999), "long"},
		{"float", NewValueFloat(3.14), "float"},
		{"float3", NewValueFloat3(1, 2, 3), "float3"},
		{"floatQ", NewValueFloatQ(0, 0, 0, 1), "floatQ"},
		{"reference", NewValueReference("ID2300"), "reference"},
		{"color", NewValueColor(1, 0, 0, 1), "color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			var check map[string]interface{}
			if err := json.Unmarshal(data, &check); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}

			if typeField, ok := check["$type"]; !ok {
				t.Error("Missing $type field")
			} else if typeField != tt.expected {
				t.Errorf("$type = %v, want %v", typeField, tt.expected)
			}

			t.Logf("%s serialized to: %s", tt.name, string(data))
		})
	}
}
