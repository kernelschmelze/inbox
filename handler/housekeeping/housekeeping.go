package housekeeping

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/kernelschmelze/inbox/model"

	log "github.com/kernelschmelze/pkg/logger"

	"github.com/google/uuid"
)

type Config struct {
	Days int
	Path string
}

type Plugin struct {
	kill chan struct{}
}

func New(config Config) *Plugin {

	plugin := &Plugin{
		kill: make(chan struct{}),
	}

	go plugin.run(config)

	return plugin
}

func (p *Plugin) Close() {
	close(p.kill)
}

func (p *Plugin) Process(inbox model.Inbox) {

}

func (p *Plugin) run(config Config) {

	keep := time.Duration(config.Days) * time.Hour * 24

	interval := time.Second // run on startup

	for {

		select {

		case <-p.kill:
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
						log.Infof("%T remove outdated file '%s'", p, name)
					} else {
						log.Errorf("%T remove outdated file '%s' failed, err=%s", p, name, err)
					}

				}

			}

			interval = 24 * time.Hour // regular interval

		}

	}

}
