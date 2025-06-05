package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	logger "github.com/sirupsen/logrus"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {

	i, err := strconv.ParseInt(BuildTime, 10, 64)
	if err == nil {
		tm := time.Unix(i, 0)
		fmt.Printf("Buildtime: ")
		fmt.Println(tm)
	} else {
		fmt.Println("BuildTime: " + BuildTime)
	}
	fmt.Println("GitCommit: " + GitCommit)
	fmt.Println("Version: " + Version)

	// terminationSignals are signals that cause the program to exit in the
	// supported platforms (linux, darwin, windows).
	terminationSignals := []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}

	c := make(chan os.Signal, 1)
	signal.Notify(c, terminationSignals...)

	go func() {
		<-c
		os.Exit(2)
	}()

	cfg := newConfig()

	switch strings.ToLower(cfg.LogLevel) {
	case "info":
		logger.SetLevel(logger.InfoLevel)
	case "warn":
		logger.SetLevel(logger.WarnLevel)
	case "error":
		logger.SetLevel(logger.ErrorLevel)
	case "fatal":
		logger.SetLevel(logger.FatalLevel)
	case "panic":
		logger.SetLevel(logger.PanicLevel)
	case "trace":
		logger.SetLevel(logger.TraceLevel)
	case "debug":
		logger.SetLevel(logger.DebugLevel)
	default:
		logger.SetLevel(logger.InfoLevel)

	}

	logger.SetFormatter(&logger.JSONFormatter{
		PrettyPrint: true,
	})

	logger.Info()
	if cfg.UnsealKeys == nil {
		logger.Fatal("unseal keys not specified")
	}

	logger.Debug("Vault Unsealer starting ...")
	monitorAndUnsealVaults(cfg.Nodes, cfg.UnsealKeys, cfg.ProbeInterval)

	logger.Info("Check and if needs be - Fixing Tokens ...")
	fixTokens(cfg.Nodes, cfg.VaultNomadServerToken, cfg.VaultToken, cfg.VaultConsulConnectToken)

	fmt.Println("\nIf testing command in build directory ... Do CTRL+C to stop ...")

	for {
		// keep job alive
		time.Sleep(time.Duration(cfg.ProbeInterval) * time.Second)
	}
}
