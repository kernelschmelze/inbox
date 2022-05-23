package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/kernelschmelze/inbox/handler"
	"github.com/kernelschmelze/inbox/handler/pushover"
	"github.com/kernelschmelze/inbox/store"
	"github.com/kernelschmelze/inbox/store/json"

	log "github.com/kernelschmelze/pkg/logger"
	"github.com/kernelschmelze/pkg/srv"

	"github.com/pelletier/go-toml"
)

const (
	maxFileSize = 1024 * 1024 * 5 // 5MB
)

type config struct {
	Listen    string
	Crt       string
	Key       string
	Path      string
	Pushover  pushover.Config
	JsonStore jsonstore.Config
}

func main() {

	path := flag.String("f", "./inbox.toml", "config file")

	flag.Parse()

	// read config from file

	config, err := readConfig(*path)
	if err != nil {
		log.Errorf("read config failed, err=%s", err)
	}

	var store store.Store

	// currently only a simple json store is supported

	if store == nil {

		if len(config.JsonStore.Path) == 0 {
			config.JsonStore.Path = "./data"
		}

		if config.JsonStore.Path, err = expandPath(config.JsonStore.Path); err != nil {
			log.Errorf("get data path failed, err=%s", err)
		} else {
			log.Infof("use data path '%s'", config.JsonStore.Path)
			store = jsonstore.New(config.JsonStore)
		}

	}

	// http handler

	handler := handler.New(maxFileSize, store)

	// pushover plugin

	if len(config.Pushover.User) != 0 && len(config.Pushover.App) != 0 {
		pushover := pushover.New(config.Pushover)
		handler.AddPlugin(pushover)
	}

	// run the server

	if len(config.Listen) == 0 {
		config.Listen = ":25478"
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
		if store != nil {
			store.Close()
		}
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

func expandPath(path string) (string, error) {

	if path == "" {
		return "", nil
	}

	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", usr.HomeDir, 1)
	}

	return filepath.Abs(path)

}
