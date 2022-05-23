package model

import (
	"time"

	"github.com/google/uuid"
)

type Inbox struct {
	Time     time.Time `json:"time"`
	ID       string    `json:"id"`
	From     string    `json:"from,omitempty"`
	Subject  string    `json:"subject,omitempty"`
	Filename string    `json:"filename,omitempty"`
	Payload  []byte    `json:"payload,omitempty"`
}

func NewInbox() (*Inbox, error) {

	inbox := &Inbox{
		Time: time.Now(),
	}

	id, err := uuid.NewRandom()
	if err == nil {
		inbox.ID = id.String()
	}

	return inbox, err

}
