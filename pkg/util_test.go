package util

import (
	"testing"
)

func TestGetAssignmentPath(t *testing.T) {
	tests := []struct {
		name       string
		channel    string
		instanceId string
		expected   string
	}{
		{
			name:       "basic path creation",
			channel:    "test-channel",
			instanceId: "instance-123",
			expected:   "test-channel.instance-123",
		},
		{
			name:       "empty channel name",
			channel:    "",
			instanceId: "instance-456",
			expected:   ".instance-456",
		},
		{
			name:       "empty instance id",
			channel:    "my-channel",
			instanceId: "",
			expected:   "my-channel.",
		},
		{
			name:       "both empty",
			channel:    "",
			instanceId: "",
			expected:   ".",
		},
		{
			name:       "special characters",
			channel:    "channel-with-dashes_and_underscores",
			instanceId: "uuid-1234-5678-9abc-def0",
			expected:   "channel-with-dashes_and_underscores.uuid-1234-5678-9abc-def0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAssignmentPath(tt.channel, tt.instanceId)
			if result != tt.expected {
				t.Errorf("GetAssignmentPath(%q, %q) = %q, want %q", tt.channel, tt.instanceId, result, tt.expected)
			}
		})
	}
}
