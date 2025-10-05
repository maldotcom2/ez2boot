package models

import "time"

type Server struct {
	Name        string `json:"name"`
	ServerGroup string `json:"server_group"`
}

type Session struct {
	Email       string    `json:"email"`
	ServerGroup string    `json:"server_group"`
	Token       string    `json:"token"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
	Message     string    `json:"message"`
}
