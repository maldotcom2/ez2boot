package notification

import (
	"encoding/json"
	"log"
	"net/http"
)

func GetNotificationTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sender, ok := GetSender("email")
		log.Print(ok)
		if ok {
			json.NewEncoder(w).Encode(sender)
		}
	}
}
