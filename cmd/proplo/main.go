package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fujiwara/proplo"
	"github.com/hashicorp/logutils"
)

func main() {
	var (
		ignore = flag.String("ignore", "", "ignore proxying network cidr")
	)
	flag.Parse()
	opt := &proplo.Options{
		LocalAddr:    flag.Args()[0],
		UpstreamAddr: flag.Args()[1],
		IgnoreCIDR:   *ignore,
	}
	opt.Validate()

	logLevel := "info"
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		logLevel = level
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	err := proplo.Run(context.Background(), opt)
	if err != nil {
		log.Println("[error]", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`proplo [local_host:port] [upstream_host:port]`)
}
