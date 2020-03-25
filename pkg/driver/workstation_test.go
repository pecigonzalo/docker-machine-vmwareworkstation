package vmwareworkstation

import (
	"errors"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

type VMWManagerMock struct {
	args   string
	stdOut string
	stdErr string
	err    error
}

func (v *VMWManagerMock) run(args ...string) error {
	_, _, err := v.runOutErr(args...)
	return err
}

func (v *VMWManagerMock) runOut(args ...string) (string, error) {
	stdout, _, err := v.runOutErr(args...)
	return stdout, err
}

func (v *VMWManagerMock) runOutErr(args ...string) (string, string, error) {
	if strings.Join(args, " ") == v.args {
		return v.stdOut, v.stdErr, v.err
	}
	return "", "", errors.New("Invalid args")
}

func newTestDriver(name string) *Driver {
	return NewDriver(name, "")
}

func TestDriverName(t *testing.T) {
	driverName := newTestDriver("default").DriverName()

	assert.Equal(t, "vmwareworkstation", driverName)
}

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}
