package vmwareworkstation

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/rancher/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]any{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestSetConfigFromFlagsShare(t *testing.T) {
	driver := NewDriver("default", "path").(*Driver)

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]any{
			"vmwareworkstation-no-share":     true,
			"vmwareworkstation-share-folder": "src:/some/dir",
			"vmwareworkstation-share-compat": "/compat/link",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	assert.NoError(t, driver.SetConfigFromFlags(checkFlags))
	assert.Empty(t, checkFlags.InvalidFlags)
	assert.True(t, driver.NoShare)
	assert.Equal(t, "src", driver.ShareName)
	assert.Equal(t, "/some/dir", driver.ShareFolder)
	assert.Equal(t, "/compat/link", driver.GuestCompatLink)
}

func TestSetConfigFromFlagsInvalidShareFolder(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]any{"vmwareworkstation-share-folder": "nocolon"},
		CreateFlags: driver.GetCreateFlags(),
	}

	assert.Error(t, driver.SetConfigFromFlags(checkFlags))
}

func TestParseShareFolder(t *testing.T) {
	tests := []struct {
		in, name, folder string
		wantErr          bool
	}{
		{in: "Users:/home/me", name: "Users", folder: "/home/me"},
		{in: `Users:C:\Users\me`, name: "Users", folder: `C:\Users\me`},
		{in: "nocolon", wantErr: true},
		{in: ":/dir", wantErr: true},
		{in: "name:", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			name, folder, err := parseShareFolder(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.name, name)
			assert.Equal(t, tt.folder, folder)
		})
	}
}

func TestSetConfigFromFlagsNetwork(t *testing.T) {
	driver := NewDriver("default", "path").(*Driver)
	opts := func(net, nic string) *drivers.CheckDriverOptions {
		return &drivers.CheckDriverOptions{
			FlagsValues: map[string]any{
				"vmwareworkstation-network-type": net,
				"vmwareworkstation-nic-type":     nic,
			},
			CreateFlags: driver.GetCreateFlags(),
		}
	}

	assert.NoError(t, driver.SetConfigFromFlags(opts("Bridged", "vmxnet3")))
	assert.Equal(t, "bridged", driver.NetworkType)
	assert.Equal(t, "vmxnet3", driver.NICType)

	assert.Error(t, driver.SetConfigFromFlags(opts("bogus", "e1000")))
	assert.Error(t, driver.SetConfigFromFlags(opts("nat", "bogus")))
}

func TestVMXNetworkTemplate(t *testing.T) {
	driver := NewDriver("default", "path").(*Driver)
	driver.NetworkType, driver.NICType = "bridged", "e1000e"
	var buf bytes.Buffer
	assert.NoError(t, template.Must(template.New("vmx").Parse(vmx)).Execute(&buf, driver))
	assert.Contains(t, buf.String(), `ethernet0.connectionType = "bridged"`)
	assert.Contains(t, buf.String(), `ethernet0.virtualDev = "e1000e"`)
}
