package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wasilak/vault_raft_snapshot_agent/config"
	"github.com/wasilak/vault_raft_snapshot_agent/snapshot_agent"
	"golang.org/x/exp/slog"
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

	slog.Info("Reading configuration...")

	c, err := config.ReadConfig()

	logLevel := "info"
	if ll := os.Getenv("VRSA_LOGLEVEL"); ll != "" {
		logLevel = ll
	}

	logFormat := "plain"
	if lg := os.Getenv("VRSA_LOGFORMAT"); lg != "" {
		logFormat = lg
	}

	config.InitLogging(logLevel, logFormat)

	if err != nil {
		slog.Error("Configuration could not be found")
		os.Exit(1)
	}

	snapshotter, err := snapshot_agent.NewSnapshotter(c)
	if err != nil {
		slog.Error("Cannot instantiate snapshotter.", slog.AnyValue(err))
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
				slog.Error(err.Error())
			} else {
				slog.Info(result)
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
			slog.Error(err.Error())
			os.Exit(1)
		} else {
			slog.Info(result)
		}
	}
}
