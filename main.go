package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/njfanxun/istio-falcon/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	command := cmd.InitCommand()
	if err := command.Execute(); err != nil {
		os.Exit(255)
	}
}
