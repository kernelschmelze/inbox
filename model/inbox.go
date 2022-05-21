package model

import (
	"time"
)

type Inbox struct {
	Time     time.Time `json:"time"`
	ID       string    `json:"id"`
	From     string    `json:"from,omitempty"`
	Subject  string    `json:"subject,omitempty"`
	Filename string    `json:"filename,omitempty"`
	Base64   bool      `json:"base64,omitempty"`
	Payload  []byte    `json:"payload,omitempty"`
}
