package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fujiwara/proplo"
)

func main() {
	if len(os.Args) != 3 {
		usage()
		return
	}
	err := proplo.Run(context.Background(), os.Args[1], os.Args[2])
	if err != nil {
		log.Println("[error]", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`proplo [local_host:port] [upstream_host:port]`)
}
