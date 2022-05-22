package pushover

import (
	"fmt"
	"time"

	"github.com/kernelschmelze/inbox/model"

	log "github.com/kernelschmelze/pkg/logger"

	po "github.com/gregdel/pushover"
)

type Config struct {
	User string
	App  string
}

type Plugin struct {
	data      chan model.Inbox
	kill      chan struct{}
}

func New(config Config) *Plugin {

	plugin := &Plugin{
		data: make(chan model.Inbox, 50),
		kill: make(chan struct{}),
	}

	go plugin.run(config)

	return plugin
}

func (p *Plugin) Close() {
	close(p.kill)
}

func (p *Plugin) Process(inbox model.Inbox) {

	select {
	case p.data <- inbox:
	case <-time.After(1 * time.Second):
		log.Errorf("%T data channel full, drop message", p)
	}

}

func (p *Plugin) run(config Config) {

	app := po.New(config.App)
	recipient := po.NewRecipient(config.User)

	for {

		select {

		case <-p.kill:
			return

		case inbox := <-p.data:

			if app == nil || recipient == nil {
				continue
			}

			title := getTitle(inbox.ID, inbox.Subject, inbox.From)

			var msg string

			if len(inbox.From) > 0 {
				msg = fmt.Sprintf("From: %s\n", inbox.From)
			}
			if len(inbox.Subject) > 0 {
				msg = msg + fmt.Sprintf("Subject: %s\n", inbox.Subject)
			}

			if len(inbox.Payload) > 0 {
				if inbox.Base64 && len(inbox.Payload)+len(msg)+3 > po.MessageMaxLength {
					msg = msg + fmt.Sprintf("\n%s\n", "payload size exceeds the limit")
				} else {
					msg = msg + fmt.Sprintf("\n%s\n", inbox.Payload)
				}
			}

			if len(msg) > po.MessageMaxLength {
				msg = msg[:po.MessageMaxLength]
			}

			message := po.NewMessageWithTitle(msg, title)
			message.Timestamp = inbox.Time.Unix()

			_, err := app.SendMessage(message, recipient)

			if err != nil {
				log.Errorf("%T post message failed, err=%s", p, err)
			}

		}

	}

}

func getTitle(names ...string) string {
	for i := range names {
		title := names[i]
		if len(title) > 0 {
			return title
		}
	}
	return ""
}
