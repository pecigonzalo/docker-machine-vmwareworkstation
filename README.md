# Docker Machine VMware Workstation Driver

[![Join the chat at https://gitter.im/pecigonzalo/docker-machine-vmwareworkstation](https://badges.gitter.im/pecigonzalo/docker-machine-vmwareworkstation.svg)](https://gitter.im/pecigonzalo/docker-machine-vmwareworkstation?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![CI](https://github.com/pecigonzalo/docker-machine-vmwareworkstation/actions/workflows/ci.yml/badge.svg)](https://github.com/pecigonzalo/docker-machine-vmwareworkstation/actions/workflows/ci.yml)

This plugin for [Docker Machine](https://docs.docker.com/machine/) creates
Docker hosts locally on a [VMware
Workstation](https://www.vmware.com/products/workstation).

This is a placeholder and collaboration point to add a VMware workstation
driver for Docker Machine. This driver reuses part of the code from the [fusion
driver](https://github.com/docker/machine/tree/master/drivers/vmwarefusion)
bundled with Docker Machine (as both have the same executable) and includes
additional code from [Packer](https://packer.io) VMware driver to detect the
location of the files on Windows systems.

This is still a work-in-progress (WIP). I'm working to add the functionality
listed on the TODO list. Suggestions and contributions are welcome.

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
* [Docker Machine](https://docs.docker.com/machine/) 0.5.0+ (built against [rancher/machine](https://github.com/rancher/machine), the maintained fork)
* [VMware Workstation](https://www.vmware.com/products/workstation) Workstation Free/Pro 10 +

The driver finds `vmrun.exe` and `vmware-vdiskmanager.exe` on your `PATH`, in
`VMWARE_HOME`, in the install directory recorded in the registry, or in the
default `Program Files` locations. If VMware Workstation is installed somewhere
else, set `VMWARE_HOME` to its install directory (and `VMWARE_DATA` to the
directory holding `vmnetdhcp.leases` if that is non-standard). `docker-machine
create` fails immediately if the tools cannot be found.

## Installation

The latest version of `docker-machine-driver-vmwareworkstation` binary is
available on the
["Releases"](https://github.com/pecigonzalo/docker-machine-vmwareworkstation/releases)
page.

Place the executable in the directory containing `docker-machine.exe`, or else
add it to your $PATH.

## Usage

Official documentation for Docker Machine is available
[here](https://docs.docker.com/machine/).

To create a VMware Workstation based Docker machine, just run this
command:

```bash
$ docker-machine create --driver=vmwareworkstation dev
```

### Docker engine options

Engine settings such as a registry mirror are handled by Docker Machine itself,
not by this driver, so they are passed to `create` directly:

```bash
$ docker-machine create --driver=vmwareworkstation \
    --engine-registry-mirror=https://my-mirror.example.com dev
```

`--engine-registry-mirror` can also be set with the `ENGINE_REGISTRY_MIRROR`
environment variable, which is useful when `create` is called from a script you
do not want to edit. `--engine-install-url` reads `MACHINE_DOCKER_INSTALL_URL`.
Other engine flags (`--engine-insecure-registry`, `--engine-opt`,
`--engine-label`, `--engine-storage-driver`, `--engine-env`) have no environment
variable and must be given on the command line.

## Options

 - `--vmwareworkstation-boot2docker-url`: The URL of the [Boot2Docker](https://github.com/boot2docker/boot2docker) image.
 - `--vmwareworkstation-disk-size`: Size of disk for the host VM (in MB).
 - `--vmwareworkstation-memory-size`: Size of memory for the host VM (in MB).
 - `--vmwareworkstation-cpu-count`: Number of CPUs to use to create the VM (-1 to use the number of CPUs available).
 - `--vmwareworkstation-ssh-user`: SSH user
 - `--vmwareworkstation-ssh-password`: SSH password
 - `--vmwareworkstation-no-share`: Disable the mount of your home directory
 - `--vmwareworkstation-share-folder`: Mount the specified directory instead of the default home location. Format: name:dir
 - `--vmwareworkstation-share-compat`: Additional link to the shared mount in the guest

The `--vmwareworkstation-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine checks locally for a
Boot2Docker ISO. If one is found, that will be used as the ISO for the new
machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the
"cached" Boot2Docker ISO.

This is the default behavior (when `--vmwareworkstation-boot2docker-url=""`),
but the option also supports specifying ISOs by the `http://` and `file://`
protocols.

Environment variables and default values:

| CLI option                            | Environment variable          | Default                  |
|---------------------------------------|-------------------------------|--------------------------|
| `--vmwareworkstation-boot2docker-url` | `WORKSTATION_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwareworkstation-network-type`    | `WORKSTATION_NETWORK_TYPE`    | `nat` (`bridged`, `hostonly`; IP discovery only works for `nat`) |
| `--vmwareworkstation-nic-type`        | `WORKSTATION_NIC_TYPE`        | `e1000` (`e1000e`, `vmxnet3`) |
| `--vmwareworkstation-cpu-count`       | `WORKSTATION_CPU_COUNT`       | `1`                      |
| `--vmwareworkstation-disk-size`       | `WORKSTATION_DISK_SIZE`       | `20000`                  |
| `--vmwareworkstation-memory-size`     | `WORKSTATION_MEMORY_SIZE`     | `1024`                   |
| `--vmwareworkstation-ssh-user`        | `WORKSTATION_SSH_USER`        | `docker`                 |
| `--vmwareworkstation-ssh-password`    | `WORKSTATION_SSH_PASSWORD`    | `tcuser`                 |
| `--vmwareworkstation-no-share`        | `WORKSTATION_NO_SHARE`        | `false`                  |
| `--vmwareworkstation-share-folder`    | `WORKSTATION_SHARE_FOLDER`    | Linux: `/home` Windows: `C:\Users\` |
| `--vmwareworkstation-share-compat`    | `WORKSTATION_SHARE_COMPAT`    | Windows: `/c/Users` |

## Development

### Build from Source

If you wish to work on VMware Workstation Driver for Docker machine, you'll
first need:

* [Go](https://go.dev/dl/) installed (the version in `go.mod` or newer).
  Alternatively, with [Nix](https://nixos.org/) run `nix develop` (or use
  `direnv allow`) to get Go, golangci-lint, GoReleaser and make at compatible
  versions.
  `make build` uses GoReleaser, which is also provided by the Nix shell.

* Currently, the build only works on Windows (WIP to get it to work on
  other platforms). Other platforms can cross-compile with `GOOS=windows`,
  which the Makefile sets by default.

To build the plugin executable binary, run these commands:

```bash
git clone https://github.com/pecigonzalo/docker-machine-vmwareworkstation
cd docker-machine-vmwareworkstation
make
```

The build creates the binary as `bin/docker-machine-driver-vmwareworkstation.exe`.
Copy it to a directory on your `PATH`.

Other useful targets: `make lint` (requires
[golangci-lint](https://golangci-lint.run/) v2; the Nix shell provides it)
and `make test`. Releases are
built by [GoReleaser](https://goreleaser.com/) when a `v*` tag is pushed.

## Authors

* Gonzalo Peci ([@pecigonzalo](https://github.com/pecigonzalo))

## Credits

* Partial copy of the README from https://github.com/Parallels/docker-machine-parallels
* [Packer](https://packer.io) VMware Workstation driver functions
