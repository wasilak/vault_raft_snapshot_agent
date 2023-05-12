package main

import (
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
	config.InitLogging()

	config.Logger.Info("Reading configuration...")

	c, err := config.ReadConfig()

	if err != nil {
		config.Logger.Error("Configuration could not be found")
		os.Exit(1)
	}

	snapshotter, err := snapshot_agent.NewSnapshotter(c)
	if err != nil {
		config.Logger.Error("Cannot instantiate snapshotter.", err)
		os.Exit(1)
	}

	if c.Daemon {
		done := listenForInterruptSignals()

		frequency, err := time.ParseDuration(c.Frequency)

		if err != nil {
			frequency = time.Hour
		}

		for {
			result, err := snapshot_agent.RunBackup(snapshotter, c)
			if err != nil {
				config.Logger.Info(err.Error())
			} else {
				config.Logger.Info(result)
			}

			select {
			case <-time.After(frequency):
				continue
			case <-done:
				os.Exit(1)
			}
		}
	} else {
		result, err := snapshot_agent.RunBackup(snapshotter, c)
		if err != nil {
			config.Logger.Error(err.Error())
			os.Exit(1)
		} else {
			config.Logger.Info(result)
		}
	}
}
