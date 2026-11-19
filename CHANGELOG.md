# Changelog

## Unreleased
FEATURES:
* Add `--vmwareworkstation-network-type` and `--vmwareworkstation-nic-type` (#27)

CHANGES:
* Default the NIC model to `e1000` instead of `vmxnet3`, which boot2docker 18.09 does not detect (#66)

IMPROVEMENTS:
* Fail fast in `create` when `vmrun`/`vmware-vdiskmanager` cannot be found, and report it from `GetState` instead of showing the machine as stopped (#46, #68)
* Document `VMWARE_HOME`/`VMWARE_DATA` for non-standard VMware installs (#68)
* Remove the Docker Toolbox `start.sh` instructions from the README; Toolbox is archived (#31, #33)
* Migrate from `dep` to Go modules and update to current Go
* Switch from the archived `docker/machine` to the maintained `github.com/rancher/machine` and update `golang.org/x/crypto`
* Read the registry with `golang.org/x/sys/windows/registry` instead of raw `syscall`
* Replace Travis CI and AppVeyor with GitHub Actions (CI and tag-based release)

## 2.0.0
IMPROVEMENTS:
* Update github.com/docker/machine to `0.14.0`

## 1.1.0
IMPROVEMENTS:
* Updated docker-machine deps
* Updated testImport deps
* Fixes #32 (Unable to complete build when .docker folder is on a symlink)

## 1.0.10
IMPROVEMENTS:
 * Mount on both /c/Users and /Users for compatibility
 * Use glide for dependancy vendoring
 * Update documentation

## 1.0.0

IMPROVEMENTS:
 * Add appveyor auto builds
 * Initial basic tests

Notes: Updated license files/notices

## 0.9.0

IMPROVEMENTS:
 * Windows: dhcp lease file is automatically discoved and not attached to a fixed location
 * Update files so it can BUILD against latest docker/machine
 * Mimic updates from docker/machine PR docker/machine#2732 and docker/machine#2721

## 0.8.0 (2015-11-17)

Initial public release
