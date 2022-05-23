package store

import (
	"github.com/kernelschmelze/inbox/model"
)

type Store interface {
	Get(id string) (*model.Inbox, error)
	Set(i *model.Inbox) (string, error)
	Close()
}
