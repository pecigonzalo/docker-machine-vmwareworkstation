package vmwareworkstation

import "os/exec"

// detect the vmrun and vmware-vdiskmanager cmds' path if needed
func detectCmdInPath(cmd string) string {
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}
