# Docker Machine VMWare Workstation Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing
to create Docker hosts locally on [VMWare Workstation](https://www.vmware.com/products/workstation)

This as placeholder and collaboration point to add a vmware workstation driver.
This is reusing part of the code from the fusion driver bundled with docker machine(as both have the same executable) and adding some code from [packer](https://packer.io) vmware driver to detect the location of the files on windows.

As the title says this is still WIP and im working to add functionality listed on the TODO list, i accept suggestions.


## TODO

* ~~drivers/vmwareworkstation/workstation.go: Rework file for vmware workstation~~
* ~~add windows support~~
* ~~add cmd/machine-driver-vmwareworkstation.go~~
* Add Linux/OSX support
* Add dhcplease file discovery on windows
* Add tests cases
* ~~Create makefile~~
* Add docs/drivers/vm-workstation.md

## Requirements
* Windows 7+ (for now)
* [Docker Machine](https://docs.docker.com/machine/) 0.5.0+
* [VMWare Workstation](https://www.vmware.com/products/workstation)

## Installation

The latest version of `docker-machine-driver-vmwareworkstation` binary is available on
the ["Releases"](https://github.com/pecigonzalo/docker-machine-vmwareworkstation/releases) page.

## Usage
Official documentation for Docker Machine [is available here](https://docs.docker.com/machine/).

To create a VMWare Workstation virtual machine for Docker purposes just run this
command:

```
$ docker-machine create --driver=vmwareworkstation dev
```

Available options:

 - `--vmwareworkstation-boot2docker-url`: The URL of the boot2docker image.
 - `--vmwareworkstation-disk-size`: Size of disk for the host VM (in MB).
 - `--vmwareworkstation-memory-size`: Size of memory for the host VM (in MB).
 - `--vmwareworkstation-cpu-count`: Number of CPUs to use to create the VM (-1 to use the number of CPUs available).
 - `--vmwareworkstation-ssh-user`: SSH user
 - `--vmwareworkstation-ssh-password`: SSH password

The `--vmwareworkstation-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO. If one is found, that will be used as the ISO for the
created machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--vmwareworkstation-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.

Environment variables and default values:

| CLI option                            | Environment variable          | Default                  |
|---------------------------------------|-------------------------------|--------------------------|
| `--vmwareworkstation-boot2docker-url` | `WORKSTATION_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwareworkstation-cpu-count`       | `WORKSTATION_CPU_COUNT`       | `1`                      |
| `--vmwareworkstation-disk-size`       | `WORKSTATION_DISK_SIZE`       | `20000`                  |
| `--vmwareworkstation-memory-size`     | `WORKSTATION_MEMORY_SIZE`     | `1024`                   |
| `--vmwareworkstation-ssh-user`        | `WORKSTATION_SSH_USER`        | `B2DUser`                |
| `--vmwareworkstation-ssh-password`    | `WORKSTATION_SSH_PASSWORD`    | `B2DPass`                |

## Development

### Build from Source
If you wish to work on VMWare Workstation Driver for Docker machine, you'll first need
* [Go](http://www.golang.org) installed (version 1.5+ is required).
  * Make sure Go is properly installed, including setting up a [GOPATH](http://golang.org/doc/code.html#GOPATH).
* [MSYS](https://msys2.github.io/)
  * **Make** We well need to use pacman to install make
* Currently build only works on Windows (WIP to get ti work on other platforms)

Run these commands to build the plugin binary:

```bash
$ go get -d github.com/pecigonzalo/docker-machine-vmwareworkstation
$ cd $GOPATH/github.com/pecigonzalo/docker-machine-vmwareworkstation
$ make
```

After the build is complete, `bin/docker-machine-driver-vmwareworkstation` binary will
be created. If you want, copy it to the `${GOPATH}/bin/`.


## Authors

* Gonzalo Peci ([@pecigonzalo](https://github.com/pecigonzalo))

## Credit

* Partial copy of the README from https://github.com/Parallels/docker-machine-parallels