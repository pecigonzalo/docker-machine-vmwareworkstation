package main

import (
	vmwareworkstation "github.com/pecigonzalo/docker-machine-vmwareworkstation"
	"github.com/rancher/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(vmwareworkstation.NewDriver("", ""))
}
