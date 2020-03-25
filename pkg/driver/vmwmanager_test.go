package vmwareworkstation

import (
	"testing"

	"os/exec"

	"errors"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestVMWMOutErr(t *testing.T) {
	var cmdRun *exec.Cmd
	vMWManager := NewVMWManager()
	vMWManager.runCmd = func(cmd *exec.Cmd) error {
		cmdRun = cmd
		fmt.Fprint(cmd.Stdout, "Printed to StdOut")
		fmt.Fprint(cmd.Stderr, "Printed to StdErr")
		return nil
	}

	stdOut, stdErr, err := vMWManager.runOutErr(vmRunBin, "arg1", "arg2")

	assert.Equal(t, []string{vmRunBin, "arg1", "arg2"}, cmdRun.Args)
	assert.Equal(t, "Printed to StdOut", stdOut)
	assert.Equal(t, "Printed to StdErr", stdErr)
	assert.NoError(t, err)
}

func TestVMWMOutErrError(t *testing.T) {
	vMWManager := NewVMWManager()
	vMWManager.runCmd = func(cmd *exec.Cmd) error { return errors.New("BUG") }

	_, _, err := vMWManager.runOutErr(vmRunBin, "arg1", "arg2")

	assert.EqualError(t, err, "BUG")
}

func TestVMWMOutErrRetryOnce(t *testing.T) {
	var cmdRun *exec.Cmd
	var runCount int
	vMWManager := NewVMWManager()
	vMWManager.runCmd = func(cmd *exec.Cmd) error {
		cmdRun = cmd

		runCount++
		if runCount == 1 {
			fmt.Fprint(cmd.Stderr, "error: The object is not ready")
			return errors.New("Fail the first time it's called")
		}

		fmt.Fprint(cmd.Stdout, "Printed to StdOut")
		return nil
	}

	stdOut, stdErr, err := vMWManager.runOutErr(vmRunBin, "command", "arg")

	assert.Equal(t, 2, runCount)
	assert.Equal(t, []string{vmRunBin, "command", "arg"}, cmdRun.Args)
	assert.Equal(t, "Printed to StdOut", stdOut)
	assert.Empty(t, stdErr)
	assert.NoError(t, err)
}
