package notification

import (
	"encoding/json"
	"ez2boot/internal/shared"
	"log"
	"net/http"
)

func GetNotificationTypes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sender, ok := GetSender("email") // Testing
		log.Print(ok)
		if ok {
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: sender}) // This might not work here
		}
	}
}
