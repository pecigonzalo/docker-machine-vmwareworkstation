package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/pecigonzalo/docker-machine-vmwareworkstation"
)

func main() {
	plugin.RegisterDriver(vmwareworkstation.NewDriver("", ""))
}
