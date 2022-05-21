package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kernelschmelze/inbox/handler"
	"github.com/kernelschmelze/inbox/handler/pushover"

	log "github.com/kernelschmelze/pkg/logger"
	"github.com/kernelschmelze/pkg/srv"

	"github.com/pelletier/go-toml"
)

const (
	maxFileSize = 1024 * 1024 * 5 // 5MB
)

type config struct {
	Listen   string
	Crt      string
	Key      string
	Pushover pushover.Config
}

func main() {

	path := flag.String("f", "./inbox.toml", "config file")

	flag.Parse()

	config, err := readConfig(*path)
	if err != nil {
		log.Errorf("read config failed, err=%s", err)
	}

	handler := handler.New("./data", maxFileSize)

	if len(config.Pushover.User) != 0 && len(config.Pushover.App) != 0 {
		pushover := pushover.New(config.Pushover)
		handler.AddPlugin(pushover)
	}

	server := srv.New(onListen, onShutdown)

	err = server.Add(srv.Config{
		config.Listen,
		handler,
		config.Crt,
		config.Key,
	})

	if err != nil {
		log.Errorf("server '%s' failed, err=%s", config.Listen, err)
	}

	defer func() {
		server.Close()
		handler.Close()
	}()

	signalHandler()

}

func signalHandler() {

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, os.Interrupt)

	select {
	case <-gracefulStop:
		fmt.Println("")
	}

	go func() {
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("kill app")
			os.Exit(1)
		}
	}()

}

func onListen(addr string, crtFile string, keyFile string) {

	if len(crtFile) == 0 || len(keyFile) == 0 {
		log.Infof("listen on '%s'", addr)
	} else {
		log.Infof("listen on '%s', crt '%s', key '%s'", addr, crtFile, keyFile)
	}

}

func onShutdown(addr string, err error) {

	if err == nil {
		log.Infof("shutdown '%s'", addr)
		return
	}

	if err != http.ErrServerClosed {
		log.Errorf("shutdown '%s', err=%s", addr, err)
	} else {
		log.Infof("shutdown '%s', err=%s", addr, err)
	}

}

func readConfig(path string) (config, error) {

	var err error
	var tml *toml.Tree

	c := config{}

	if tml, err = toml.LoadFile(path); err == nil {
		err = tml.Unmarshal(&c)
	}

	return c, err
}
