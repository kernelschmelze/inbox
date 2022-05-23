package model

type Plugin interface {
	Process(*Inbox)
	Close()
}
