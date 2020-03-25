package vmwareworkstation

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
)

const (
	retryCountOnError = 5
	retryDelay        = 100 * time.Millisecond
)

var (
	ErrMachineExist    = errors.New("machine already exists")
	ErrMachineNotExist = errors.New("machine does not exist")
	ErrVMRUNNotFound   = errors.New("VMRUN not found")
)

var (
	vmRunBin                  = detectVMRunCmd()
	vDiskManBin               = detectVDiskManCmd()
	workstationDhcpLeasesPath = detectDhcpLeasesPath()
)

// VMWManager defines a VMWare manager interface
type VMWManager interface {
	run(bin string, args ...string) error

	runOut(bin string, args ...string) (string, error)

	runOutErr(bin string, args ...string) (string, string, error)
}

// VMWCmdManager communicates with VMWare through the commandline using `vmrun`.
type VMWCmdManager struct {
	runCmd func(cmd *exec.Cmd) error
}

// NewVMWManager creates a VBoxManager instance.
func NewVMWManager() *VMWCmdManager {
	return &VMWCmdManager{
		runCmd: func(cmd *exec.Cmd) error { return cmd.Run() },
	}
}

func (v *VMWCmdManager) run(bin string, args ...string) error {
	_, _, err := v.runOutErr(bin, args...)
	return err
}

func (v *VMWCmdManager) runOut(bin string, args ...string) (string, error) {
	stdout, _, err := v.runOutErr(bin, args...)
	return stdout, err
}

func (v *VMWCmdManager) runOutErr(bin string, args ...string) (string, string, error) {
	return v.runOutErrRetry(bin, retryCountOnError, args...)
}

func (v *VMWCmdManager) runOutErrRetry(bin string, retry int, args ...string) (string, string, error) {
	cmd := exec.Command(bin, args...)
	log.Debugf("executing: %v %v", bin, strings.Join(args, " "))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := v.runCmd(cmd)
	stderrStr := stderr.String()
	if os.Getenv("MACHINE_DEBUG") != "" {
		log.Debugf("stdout:\n{\n%v}", stdout.String())
		log.Debugf("stderr:\n{\n%v}", stderrStr)
	}
	if err != nil {
		if ee, ok := err.(*exec.Error); ok && ee == exec.ErrNotFound {
			err = ErrVMRUNNotFound
		}
	}

	// Sometimes, we just need to retry...
	if retry > 1 {
		if err != nil {
			time.Sleep(retryDelay)
			return v.runOutErrRetry(bin, retry-1, args...)
		}
	}

	return stdout.String(), stderrStr, err
}
