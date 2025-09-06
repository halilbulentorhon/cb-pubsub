package util

import (
	"fmt"
)

func GetAssignmentPath(groupName, instanceId string) string {
	return fmt.Sprintf("%s.%s", groupName, instanceId)
}
