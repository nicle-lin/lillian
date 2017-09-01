package model

import "time"

type Event struct {
	Type     string    `json:"type,omitempty"`
	Time     time.Time `json:"time,omitempty"`
	Message  string    `json:"message,omitempty"`
	Username string    `json:"username,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
}
