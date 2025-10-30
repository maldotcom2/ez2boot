package notification

import "slices"

var SupportedNotificationTypes = []string{"email"}

// No enums :(
func IsSupportedNotificationType(nt string) bool {
	return slices.Contains(SupportedNotificationTypes, nt)
}
