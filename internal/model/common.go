package model

import "time"

type Config struct {
	CloudProvider  string
	Port           string
	ScrapeInterval time.Duration
	TagKey         string
	// Add more fields as needed
}

type Server struct {
	UniqueID    string `json:"unique_id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	ServerGroup string `json:"server_group"`
	TimeAdded   int64  `json:"time_added"`
}

type Session struct {
	Email       string    `json:"email"`
	ServerGroup string    `json:"server_group"`
	Token       string    `json:"token"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
	Message     string    `json:"message"`
}
