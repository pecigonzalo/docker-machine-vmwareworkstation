# Docker Machine VMware Workstation Driver

[![Join the chat at https://gitter.im/pecigonzalo/docker-machine-vmwareworkstation](https://badges.gitter.im/pecigonzalo/docker-machine-vmwareworkstation.svg)](https://gitter.im/pecigonzalo/docker-machine-vmwareworkstation?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/k8j7ej2a7t58p2r0/branch/master?svg=true)](https://ci.appveyor.com/project/pecigonzalo/docker-machine-vmwareworkstation)

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing
to create Docker hosts locally on [VMware Workstation](https://www.vmware.com/products/workstation)

This as placeholder and collaboration point to add a vmware workstation driver.
This is reusing part of the code from the fusion driver bundled with docker machine(as both have the same executable) and adding some code from [Packer](https://packer.io) vmware driver to detect the location of the files on windows.

As the title says this is still WIP and im working to add functionality listed on the TODO list, i accept suggestions.


## TODO

* ~~drivers/vmwareworkstation/workstation.go: Rework file for vmware workstation~~
* ~~add windows support~~
* ~~add cmd/machine-driver-vmwareworkstation.go~~
* Add Linux/OSX support
* ~~Add dhcplease file discovery on windows~~
* Add tests cases
* ~~Create makefile~~
* Add docs/drivers/vm-workstation.md

## Requirements
* Windows 7+ (for now)
* [Docker Machine](https://docs.docker.com/machine/) 0.5.0+
* [VMware Workstation](https://www.vmware.com/products/workstation) Workstation Free/Pro 10 +

## Installation

The latest version of `docker-machine-driver-vmwareworkstation` binary is available on
the ["Releases"](https://github.com/pecigonzalo/docker-machine-vmwareworkstation/releases) page.

Put the executable in the directory containing docker-machine.exe, or else add it to your $PATH.

## Installing with Docker Toolbox
InstallDocker Toolbox without VirtualBox

`DockerToolbox-.exe /COMPONENTS="Docker,DockerMachine"`

Replace contents of "C:\Program Files\Docker Toolbox\start.sh" with this

```bash
#!/bin/bash

export PATH="$PATH:/c/Program Files (x86)/VMware/VMware Workstation"

trap '[ "$?" -eq 0 ] || read -p "Looks like something went wrong in step ´$STEP´... Press any key to continue..."' EXIT

VM=${DOCKER_MACHINE_NAME-default}
DOCKER_MACHINE=./docker-machine.exe

BLUE='\033[1;34m'
GREEN='\033[0;32m'
NC='\033[0m'


if [ ! -f "${DOCKER_MACHINE}" ]; then
  echo "Docker Machine is not installed. Please re-run the Toolbox Installer and try again."
  exit 1
fi

vmrun.exe list | grep \""${VM}"\" &> /dev/null
VM_EXISTS_CODE=$?

set -e

STEP="Checking if machine $VM exists"
if [ $VM_EXISTS_CODE -eq 1 ]; then
  "${DOCKER_MACHINE}" rm -f "${VM}" &> /dev/null || :
  rm -rf ~/.docker/machine/machines/"${VM}"
  #set proxy variables if they exists
  if [ -n ${HTTP_PROXY+x} ]; then
    PROXY_ENV="$PROXY_ENV --engine-env HTTP_PROXY=$HTTP_PROXY"
  fi
  if [ -n ${HTTPS_PROXY+x} ]; then
    PROXY_ENV="$PROXY_ENV --engine-env HTTPS_PROXY=$HTTPS_PROXY"
  fi
  if [ -n ${NO_PROXY+x} ]; then
    PROXY_ENV="$PROXY_ENV --engine-env NO_PROXY=$NO_PROXY"
  fi  
  "${DOCKER_MACHINE}" create -d vmwareworkstation $PROXY_ENV "${VM}"
fi

STEP="Checking status on $VM"
VM_STATUS="$(${DOCKER_MACHINE} status ${VM} 2>&1)"
if [ "${VM_STATUS}" != "Running" ]; then
  "${DOCKER_MACHINE}" start "${VM}"
  yes | "${DOCKER_MACHINE}" regenerate-certs "${VM}"
fi

STEP="Setting env"
eval "$(${DOCKER_MACHINE} env --shell=bash ${VM})"

STEP="Finalize"
clear
cat << EOF


                        ##         .
                  ## ## ##        ==
               ## ## ## ## ##    ===
           /"""""""""""""""""\___/ ===
      ~~~ {~~ ~~~~ ~~~ ~~~~ ~~~ ~ /  ===- ~~~
           \______ o           __/
             \    \         __/
              \____\_______/

EOF
echo -e "${BLUE}docker${NC} is configured to use the ${GREEN}${VM}${NC} machine with IP ${GREEN}$(${DOCKER_MACHINE} ip ${VM})${NC}"
echo "For help getting started, check out the docs at https://docs.docker.com"
echo
cd

docker () {
  MSYS_NO_PATHCONV=1 docker.exe "$@"
}
export -f docker

if [ $# -eq 0 ]; then
  echo "Start interactive shell"
  exec "$BASH" --login -i
else
  echo "Start shell with command"
  exec "$BASH" -c "$*"
fi
```
Credit to [@gtirloni](https://github.com/gtirloni)

## Usage
Official documentation for Docker Machine [is available here](https://docs.docker.com/machine/).

To create a VMware Workstation virtual machine for Docker purposes just run this
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
| `--vmwareworkstation-ssh-user`        | `WORKSTATION_SSH_USER`        | `docker`                 |
| `--vmwareworkstation-ssh-password`    | `WORKSTATION_SSH_PASSWORD`    | `tcuser`                 |

## Development

### Build from Source
If you wish to work on VMWare Workstation Driver for Docker machine, you'll first need
* [Go](http://www.golang.org) installed (version 1.6+ is required).
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

## Credits

* Partial copy of the README from https://github.com/Parallels/docker-machine-parallels
* [Packer](https://packer.io) VMWare Workstation driver functions
* [gtirloni](https://github.com/gtirloni) Instructions for Docker Toolbox
