package main

import (
	"github/njfanxun/istio-falcon/cmd"
	"github/njfanxun/istio-falcon/internal/boot"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	boot.InitBoot()
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("%+v", err)
			os.Exit(-1)
		}
	}()
	command := cmd.InitCommand()
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}
