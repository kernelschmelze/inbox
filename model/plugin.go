package model

type Plugin interface {
	Process(i Inbox)
	Close()
}
