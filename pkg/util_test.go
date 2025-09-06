package util

import (
	"testing"
)

func TestGetAssignmentPath(t *testing.T) {
	tests := []struct {
		name       string
		groupName  string
		instanceId string
		expected   string
	}{
		{
			name:       "basic path creation",
			groupName:  "test-group",
			instanceId: "instance-123",
			expected:   "test-group.instance-123",
		},
		{
			name:       "empty group name",
			groupName:  "",
			instanceId: "instance-456",
			expected:   ".instance-456",
		},
		{
			name:       "empty instance id",
			groupName:  "my-group",
			instanceId: "",
			expected:   "my-group.",
		},
		{
			name:       "both empty",
			groupName:  "",
			instanceId: "",
			expected:   ".",
		},
		{
			name:       "special characters",
			groupName:  "group-with-dashes_and_underscores",
			instanceId: "uuid-1234-5678-9abc-def0",
			expected:   "group-with-dashes_and_underscores.uuid-1234-5678-9abc-def0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAssignmentPath(tt.groupName, tt.instanceId)
			if result != tt.expected {
				t.Errorf("GetAssignmentPath(%q, %q) = %q, want %q", tt.groupName, tt.instanceId, result, tt.expected)
			}
		})
	}
}
