/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */
/*
 * Copyright 2015 Gonzalo Peci  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwareworkstation

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"net"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/rancher/machine/libmachine/drivers"
	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/libmachine/mcnflag"
	"github.com/rancher/machine/libmachine/mcnutils"
	"github.com/rancher/machine/libmachine/ssh"
	"github.com/rancher/machine/libmachine/state"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	B2DUser        = "docker"
	B2DPass        = "tcuser"
	isoFilename    = "boot2docker.iso"
	isoConfigDrive = "configdrive.iso"

	defaultSSHUser  = B2DUser
	defaultSSHPass  = B2DPass
	defaultDiskSize = 20000
	defaultCpus     = 1
	defaultMemory   = 1024
)

// Driver for VMware Workstation
type Driver struct {
	*drivers.BaseDriver
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string
	CPUS           int

	SSHPassword    string
	ConfigDriveISO string
	ConfigDriveURL string

	NoShare         bool
	ShareName       string
	ShareFolder     string
	GuestFolder     string
	GuestCompatLink string

	mac string // cached VM MAC address, see macAddress
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_BOOT2DOCKER_URL",
			Name:   "vmwareworkstation-boot2docker-url",
			Usage:  "VMWare Workstation URL for boot2docker image",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_CONFIGDRIVE_URL",
			Name:   "vmwareworkstation-configdrive-url",
			Usage:  "VMWare Workstation URL for cloud-init configdrive",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "WORKSTATION_CPU_COUNT",
			Name:   "vmwareworkstation-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCpus,
		},
		mcnflag.IntFlag{
			EnvVar: "WORKSTATION_MEMORY_SIZE",
			Name:   "vmwareworkstation-memory-size",
			Usage:  "VMWare Workstation size of memory for host VM (in MB)",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "WORKSTATION_DISK_SIZE",
			Name:   "vmwareworkstation-disk-size",
			Usage:  "VMWare Workstation size of disk for host VM (in MB)",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_SSH_USER",
			Name:   "vmwareworkstation-ssh-user",
			Usage:  "SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_SSH_PASSWORD",
			Name:   "vmwareworkstation-ssh-password",
			Usage:  "SSH password",
			Value:  defaultSSHPass,
		},
		mcnflag.BoolFlag{
			EnvVar: "WORKSTATION_NO_SHARE",
			Name:   "vmwareworkstation-no-share",
			Usage:  "Disable the mount of your home directory",
		},
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_SHARE_FOLDER",
			Name:   "vmwareworkstation-share-folder",
			Usage:  "Mount the specified directory instead of the default home location. Format: name:dir",
		},
		mcnflag.StringFlag{
			EnvVar: "WORKSTATION_SHARE_COMPAT",
			Name:   "vmwareworkstation-share-compat",
			Usage:  "Override the compatibility link created by this driver",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		CPUS:        defaultCpus,
		Memory:      defaultMemory,
		DiskSize:    defaultDiskSize,
		SSHPassword: defaultSSHPass,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "vmwareworkstation"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwareworkstation-memory-size")
	d.CPU = flags.Int("vmwareworkstation-cpu-count")
	d.DiskSize = flags.Int("vmwareworkstation-disk-size")
	d.Boot2DockerURL = flags.String("vmwareworkstation-boot2docker-url")
	d.ConfigDriveURL = flags.String("vmwareworkstation-configdrive-url")
	d.ISO = d.ResolveStorePath(isoFilename)
	d.ConfigDriveISO = d.ResolveStorePath(isoConfigDrive)
	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = flags.String("vmwareworkstation-ssh-user")
	d.SSHPassword = flags.String("vmwareworkstation-ssh-password")
	d.SSHPort = 22

	// We support a maximum of 16 cpu to be consistent with Virtual Hardware 10
	// specs.
	if d.CPU < 1 {
		d.CPU = runtime.NumCPU()
	}
	if d.CPU > 16 {
		d.CPU = 16
	}

	d.NoShare = flags.Bool("vmwareworkstation-no-share")
	switch runtime.GOOS {

	case "linux": // TODO Test linux working
		d.ShareName = "Home"
		d.ShareFolder = "/home"
		d.GuestFolder = "/Users"
	case "windows":
		d.ShareName = "Users"
		d.ShareFolder = `C:\Users\`
		d.GuestFolder = "/Users"
		d.GuestCompatLink = "/c/Users"
	}

	if shareFolder := flags.String("vmwareworkstation-share-folder"); shareFolder != "" {
		name, folder, err := parseShareFolder(shareFolder)
		if err != nil {
			return err
		}
		d.ShareName, d.ShareFolder = name, folder
	}
	if compat := flags.String("vmwareworkstation-share-compat"); compat != "" {
		d.GuestCompatLink = compat
	}

	return nil
}

// parseShareFolder splits a "name:dir" value into the share name and the host
// directory. The directory may itself contain colons (e.g. C:\Users).
func parseShareFolder(value string) (name, folder string, err error) {
	name, folder, ok := strings.Cut(value, ":")
	if !ok || name == "" || folder == "" {
		return "", "", fmt.Errorf("invalid share folder %q, expected format name:dir", value)
	}
	return name, folder, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	ip, err := d.getIPfromDHCPLease()
	if err != nil {
		return "", err
	}

	return ip, nil
}

// GetState returns the current state of the machine.
func (d *Driver) GetState() (state.State, error) {
	// VMRUN only tells use if the vm is running or not
	if _, err := os.Stat(d.vmxPath()); os.IsNotExist(err) {
		return state.Error, err
	}
	if stdout, _, _ := vmrun("list"); strings.Contains(stdout, d.vmxPath()) {
		return state.Running, nil
	}
	return state.Stopped, nil
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.UpdateISOCache(d.Boot2DockerURL); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	// download cloud-init config drive
	if d.ConfigDriveURL != "" {
		if err := b2dutils.DownloadISO(d.ResolveStorePath("."), isoConfigDrive, d.ConfigDriveURL); err != nil {
			return err
		}
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if err := os.MkdirAll(d.ResolveStorePath("."), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(d.vmxPath()); err == nil {
		return ErrMachineExist
	}

	// Generate vmx config file from template
	vmxt := template.Must(template.New("vmx").Parse(vmx))
	var vmxbuf bytes.Buffer
	if err := vmxt.Execute(&vmxbuf, d); err != nil {
		return err
	}
	if err := os.WriteFile(d.vmxPath(), vmxbuf.Bytes(), 0o644); err != nil {
		return err
	}

	// Generate vmdk file
	diskImg := d.ResolveStorePath(fmt.Sprintf("%s.vmdk", d.MachineName))
	if _, err := os.Stat(diskImg); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := vdiskmanager(diskImg, d.DiskSize); err != nil {
			return err
		}
	}

	log.Infof("Starting %s...", d.MachineName)
	if err := vmrunChecked("start", "start", d.vmxPath(), "nogui"); err != nil {
		return err
	}

	var (
		ip  string
		err error
	)

	log.Infof("Waiting for VM to come online...")
	for i := 1; i <= 60; i++ {
		ip, err = d.getIPfromDHCPLease()
		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if ip != "" {
			log.Debugf("Got an ip: %s", ip)
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, 22), 2*time.Second)
			if err != nil {
				log.Debugf("SSH Daemon not responding yet: %s", err)
				time.Sleep(2 * time.Second)
				continue
			}
			_ = conn.Close()
			break
		}
	}

	if ip == "" {
		return fmt.Errorf("machine didn't return an IP after 120 seconds, aborting")
	}

	// we got an IP, let's copy ssh keys over
	d.IPAddress = ip

	// Do not execute the rest of boot2docker specific configuration
	// The uplaod of the public ssh key uses a ssh connection,
	// this works without installed vmware client tools
	if d.ConfigDriveURL != "" {
		var keyfh *os.File
		var keycontent []byte

		log.Infof("Copy public SSH key to %s [%s]", d.MachineName, d.IPAddress)

		// create .ssh folder in users home
		if err := executeSSHCommand(fmt.Sprintf("mkdir -p /home/%s/.ssh", d.SSHUser), d); err != nil {
			return err
		}

		// read generated public ssh key
		if keyfh, err = os.Open(d.publicSSHKeyPath()); err != nil {
			return err
		}
		defer func() { _ = keyfh.Close() }()

		if keycontent, err = io.ReadAll(keyfh); err != nil {
			return err
		}

		// add public ssh key to authorized_keys
		if err := executeSSHCommand(fmt.Sprintf("echo '%s' > /home/%s/.ssh/authorized_keys", string(keycontent), d.SSHUser), d); err != nil {
			return err
		}

		// make it secure
		if err := executeSSHCommand(fmt.Sprintf("chmod 600 /home/%s/.ssh/authorized_keys", d.SSHUser), d); err != nil {
			return err
		}

		log.Debugf("Leaving create sequence early, configdrive found")
		return nil
	}

	// Generate a tar keys bundle
	if err := d.generateKeyBundle(); err != nil {
		return err
	}

	// Test if /var/lib/boot2docker exists
	vmrunBestEffort("-gu", B2DUser, "-gp", B2DPass, "directoryExistsInGuest", d.vmxPath(), "/var/lib/boot2docker")

	// Copy SSH keys bundle
	if err := vmrunChecked("CopyFileFromHostToGuest", "-gu", B2DUser, "-gp", B2DPass, "CopyFileFromHostToGuest", d.vmxPath(), d.ResolveStorePath("userdata.tar"), "/home/docker/userdata.tar"); err != nil {
		return err
	}

	// Expand tar file.
	vmrunBestEffort("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo /bin/mv /home/docker/userdata.tar /var/lib/boot2docker/userdata.tar && sudo tar xf /var/lib/boot2docker/userdata.tar -C /home/docker/ > /var/log/userdata.log 2>&1 && sudo chown -R docker:staff /home/docker")

	if !d.NoShare {
		// Enable Shared Folders
		vmrunBestEffort("-gu", B2DUser, "-gp", B2DPass, "enableSharedFolders", d.vmxPath())
		if err := mountSharedFolder(d); err != nil {
			return err
		}
	} else {
		log.Infof("No shared folders")
	}

	return nil
}

func (d *Driver) Start() error {
	if err := vmrunChecked("start", "start", d.vmxPath(), "nogui"); err != nil {
		return err
	}

	// Do not execute the rest of boot2docker specific configuration, exit here
	if d.ConfigDriveURL != "" {
		log.Debugf("Leaving start sequence early, configdrive found")
		return nil
	}

	if !d.NoShare {
		if err := mountSharedFolder(d); err != nil {
			return err
		}
	} else {
		log.Infof("No shared folders")
	}

	return nil
}

func (d *Driver) Stop() error {
	_, _, err := vmrun("stop", d.vmxPath(), "nogui")
	return err
}

func (d *Driver) Remove() error {
	s, _ := d.GetState()
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return fmt.Errorf("error stopping VM before deletion")
		}
	}
	log.Infof("Deleting %s...", d.MachineName)
	return vmrunChecked("deleteVM", "deleteVM", d.vmxPath(), "nogui")
}

func (d *Driver) Restart() error {
	_, _, err := vmrun("reset", d.vmxPath(), "nogui")

	if !d.NoShare {
		if err := mountSharedFolder(d); err != nil {
			return err
		}
	} else {
		log.Infof("No shared folders")
	}

	return err
}

func (d *Driver) Kill() error {
	_, _, err := vmrun("stop", d.vmxPath(), "hard nogui")
	return err
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("VMware Workstation does not currently support the upgrade operation")
}

func (d *Driver) vmxPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.vmx", d.MachineName))
}

var (
	// vmxparse finds the generated MAC address, as we pass a VMX with addressType = "generated".
	vmxparse = regexp.MustCompile(`^ethernet0.generatedAddress\s*=\s*"(.*?)"\s*$`)
	// leaseip gets the IP from the lease table.
	leaseip = regexp.MustCompile(`^lease (.+?) {$`)
	// leaseend gets the lease end date time.
	leaseend = regexp.MustCompile(`^\s*ends \d (.+?);$`)
	// leasemac gets the MAC address associated.
	leasemac = regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)
)

// macAddress returns the VM's generated MAC address from its VMX file. VMware
// only writes it on first power-on, so it is looked up until found and cached.
func (d *Driver) macAddress() (string, error) {
	if d.mac != "" {
		return d.mac, nil
	}

	vmxcontent, err := os.ReadFile(d.vmxPath())
	if err != nil {
		return "", err
	}

	var macaddr string
	for line := range strings.SplitSeq(string(vmxcontent), "\n") {
		if matches := vmxparse.FindStringSubmatch(line); matches != nil {
			macaddr = strings.ToLower(matches[1])
		}
	}
	if macaddr == "" {
		return "", fmt.Errorf("couldn't find MAC address in VMX file %s", d.vmxPath())
	}

	log.Debugf("MAC address in VMX: %s", macaddr)
	d.mac = macaddr
	return macaddr, nil
}

func (d *Driver) getIPfromDHCPLease() (string, error) {
	// DHCP lease table for NAT vmnet interface
	dhcpfile := dhcpLeasesFilePath()
	if dhcpfile == "" {
		return "", fmt.Errorf("no DHCP leases path found")
	}

	macaddr, err := d.macAddress()
	if err != nil {
		return "", err
	}

	// The lease table changes while the VM boots, so it is re-read on each call.
	dhcpcontent, err := os.ReadFile(dhcpfile)
	if err != nil {
		return "", err
	}

	var (
		lastipmatch         string
		currentip           string
		lastleaseendtime    time.Time
		currentleadeendtime time.Time
	)
	for line := range strings.SplitSeq(string(dhcpcontent), "\r\n") {
		if matches := leaseip.FindStringSubmatch(line); matches != nil {
			lastipmatch = matches[1]
			continue
		}

		if matches := leaseend.FindStringSubmatch(line); matches != nil {
			lastleaseendtime, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		if matches := leasemac.FindStringSubmatch(line); matches != nil && matches[1] == macaddr && currentleadeendtime.Before(lastleaseendtime) {
			currentip = lastipmatch
			currentleadeendtime = lastleaseendtime
		}
	}

	if currentip == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases", macaddr)
	}

	log.Debugf("IP found in DHCP lease table: %s", currentip)
	return currentip, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Make a boot2docker userdata.tar key bundle
func (d *Driver) generateKeyBundle() (err error) {
	log.Debugf("Creating Tar key bundle...")

	magicString := "boot2docker, this is vmware speaking"

	tf, err := os.Create(d.ResolveStorePath("userdata.tar"))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := tf.Close(); err == nil {
			err = cerr
		}
	}()

	tw := tar.NewWriter(tf)

	// magicString first so we can figure out who originally wrote the tar.
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	pubKey, err := os.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	return tw.Close()
}

// execute command over SSH with user / password authentication
func executeSSHCommand(command string, d *Driver) error {
	log.Debugf("Execute executeSSHCommand: %s", command)

	config := &cryptossh.ClientConfig{
		User: d.SSHUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(d.SSHPassword),
		},
	}

	client, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), config)
	if err != nil {
		log.Debugf("Failed to dial: %s", err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		log.Debugf("Failed to create session: %s", err)
		return err
	}
	defer func() { _ = session.Close() }()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		log.Debugf("Failed to run: %s", err)
		return err
	}
	log.Debugf("Stdout from executeSSHCommand: %s", b.String())

	return nil
}

// vmrunChecked runs vmrun and returns an error, naming the operation and
// including vmrun's output, if it fails.
func vmrunChecked(op string, args ...string) error {
	stdout, stderr, err := vmrun(args...)
	if err != nil {
		return fmt.Errorf("vmrun %s: %w: %s", op, err, strings.TrimSpace(stdout+stderr))
	}
	return nil
}

// vmrunBestEffort runs vmrun and logs, rather than returns, any failure. It is
// only for optional guest-side steps whose failure must not abort the
// operation.
func vmrunBestEffort(args ...string) {
	if _, _, err := vmrun(args...); err != nil {
		log.Warnf("vmrun failed (continuing): %s", err)
	}
}

func mountSharedFolder(d *Driver) error {
	log.Infof("Mounting Shared Folders...")
	if d.ShareFolder == "" {
		return nil
	}

	if _, err := os.Stat(d.ShareFolder); errors.Is(err, fs.ErrNotExist) {
		log.Warnf("Shared folder %s does not exist on host, skipping", d.ShareFolder)
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot access shared folder %s: %w", d.ShareFolder, err)
	}

	// Add the share to the VM configuration, then create the mountpoint and mount it.
	log.Infof("Adding shared folder %s and mapping to /%s ...", d.ShareFolder, d.ShareName)
	vmrunBestEffort(
		"-gu", B2DUser, "-gp", B2DPass, "addSharedFolder", d.vmxPath(),
		d.ShareName,
		vmrunHostPath(d.ShareFolder),
	)

	commands := guestMountCommands(d)
	log.Debug(commands)
	for _, command := range commands {
		vmrunBestEffort("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", command)
	}
	return nil
}

// vmrunHostPath converts forward slashes to backslashes so that a host path
// such as C:/Users/me is accepted by vmrun on Windows.
func vmrunHostPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

// guestMountCommands returns the shell commands that create the mountpoint,
// mount the shared folder and, if configured, add the compatibility symlink.
func guestMountCommands(d *Driver) []string {
	commands := []string{
		fmt.Sprintf("[ ! -d %[1]q ] && sudo mkdir %[1]q", d.GuestFolder),
		fmt.Sprintf(
			"[ -f /usr/local/bin/vmhgfs-fuse ] && "+
				"sudo /usr/local/bin/vmhgfs-fuse -o allow_other .host:/%[1]v %[2]q"+
				" || "+
				"sudo mount -t vmhgfs .host:/%[1]v %[2]q",
			d.ShareName,
			d.GuestFolder,
		),
	}

	if d.GuestCompatLink != "" {
		commands = append(commands,
			fmt.Sprintf("sudo mkdir -p %q", path.Dir(d.GuestCompatLink)),
			fmt.Sprintf("[ -e %q ] || sudo ln -s %q %q", d.GuestCompatLink, d.GuestFolder, d.GuestCompatLink),
		)
	}
	return commands
}
