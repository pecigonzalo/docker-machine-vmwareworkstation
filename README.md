# docker-machine-vmwareworkstation
VMWare Workstation driver for Docker Machine https://github.com/docker/machine



Add this as placeholder and collaboration point to add a vmware workstation driver.
This is reusing part of the code from the fusion driver as both have the same executable and adding some code from packer vmware driver to detect the location of the files on windows.

As the title says this is still WIP.

TODO:

* ~~drivers/vmwareworkstation/workstation.go: Rework file for vmware workstation~~
* ~~add windows support~~
* ~~add cmd/machine-driver-vmwareworkstation.go~~
* add linux support
* add dhcplease file discovery on windows
* add tests cases
* create makefile
* add docs/drivers/vm-workstation.md
