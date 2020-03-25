package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	driver "github.com/pecigonzalo/docker-machine-vmwareworkstation/pkg/driver"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}
