package jsonstore

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/kernelschmelze/inbox/model"
)

type Config struct {
	Path string
	Days int
}

type Store struct {
	config Config
	kill   chan struct{}
}

func New(config Config) *Store {

	store := &Store{
		config: config,
		kill:   make(chan struct{}),
	}

	if config.Days > 0 {
		go store.run(config)
	}

	return store

}

func (s *Store) Get(id string) (*model.Inbox, error) {

	inbox := &model.Inbox{}

	path := path.Join(s.config.Path, id)

	buffer, err := ioutil.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(buffer, &inbox)
	}

	return inbox, err

}

func (s *Store) Set(i *model.Inbox) (string, error) {

	if i == nil {
		return "", nil
	}

	id := i.ID

	path := path.Join(s.config.Path, id)
	if err := mkdir(path); err != nil {
		return id, err
	}

	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return id, err
	}

	defer output.Close()

	encoder := json.NewEncoder(output)
	err = encoder.Encode(i)

	return id, err

}

func (s *Store) Close() {
	close(s.kill)
}

func mkdir(file string) error {
	path := file
	index := strings.LastIndex(file, "/")
	if index > 0 {
		path = file[0:index]
	}
	err := os.MkdirAll(path, 0700)
	return err
}
