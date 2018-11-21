package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/zyf0330/confd/backends"
	"github.com/zyf0330/confd/log"
	"github.com/zyf0330/confd/resource/template"
)

func main() {
	flag.Parse()
	if config.PrintVersion {
		fmt.Printf("confd %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
		os.Exit(0)
	}
	if config.PProf {
		log.Info("start pprof server at localhost:6060")
		go func() {
			log.Error(fmt.Sprintf("%s", http.ListenAndServe("localhost:6060", nil)))
		}()
	}
	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting confd")

	storeClient, err := backends.New(config.BackendsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.TemplateConfig.StoreClient = storeClient
	if config.OneTime {
		if err := template.Process(config.TemplateConfig); err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(config.TemplateConfig, stopChan, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(config.TemplateConfig, stopChan, doneChan, errChan, config.Interval)
	}

	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
