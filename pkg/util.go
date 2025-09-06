package util

import (
	"fmt"
)

func GetAssignmentPath(chanel, instanceId string) string {
	return fmt.Sprintf("%s.%s", chanel, instanceId)
}
