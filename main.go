package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wasilak/vault_raft_snapshot_agent/config"
	"github.com/wasilak/vault_raft_snapshot_agent/snapshot_agent"
)

func listenForInterruptSignals() chan bool {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		done <- true
	}()
	return done
}

func main() {
	log.Println("Reading configuration...")

	c, err := config.ReadConfig()

	if err != nil {
		log.Fatalln("Configuration could not be found")
	}

	snapshotter, err := snapshot_agent.NewSnapshotter(c)
	if err != nil {
		log.Fatalln("Cannot instantiate snapshotter.", err)
	}

	if c.Daemon {
		done := listenForInterruptSignals()

		frequency, err := time.ParseDuration(c.Frequency)

		if err != nil {
			frequency = time.Hour
		}

		for {
			err = snapshot_agent.RunBackup(snapshotter, c)
			if err != nil {
				log.Println(err)
			}

			select {
			case <-time.After(frequency):
				continue
			case <-done:
				os.Exit(1)
			}
		}
	} else {
		err = snapshot_agent.RunBackup(snapshotter, c)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
