package jsonstore

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/google/uuid"
)

func (s *Store) run(config Config) {

	keep := time.Duration(config.Days) * time.Hour * 24

	interval := time.Second // run on startup

	for {

		select {

		case <-s.kill:
			return

		case <-time.After(interval):

			fileInfo, err := ioutil.ReadDir(config.Path)
			if err != nil {
				interval = 1 * time.Minute // retry on error
				continue
			}

			for _, file := range fileInfo {

				if file.IsDir() {
					continue
				}

				deadline := file.ModTime().Add(keep)

				if deadline.Before(time.Now()) {

					if _, err := uuid.Parse(file.Name()); err != nil {
						continue
					}

					name := path.Join(config.Path, file.Name())

					if err := os.Remove(name); err == nil {
						log.Infof("%T remove outdated file '%s'", s, name)
					} else {
						log.Errorf("%T remove outdated file '%s' failed, err=%s", s, name, err)
					}

				}

			}

			interval = 1 * time.Hour // regular interval

		}

	}

}
