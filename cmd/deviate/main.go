package main

import (
	"github.com/openshift-knative/deviate/internal/cmd"
	"github.com/wavesoftware/go-commandline"
)

func main() {
	commandline.New(new(cmd.App)).ExecuteOrDie(cmd.Options...)
}
