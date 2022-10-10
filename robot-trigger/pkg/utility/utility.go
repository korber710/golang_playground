package utility

import (
	"os"
	"os/exec"
)

func ExecShell(commandName string, commandList []string) error {
	cmd := exec.Command(commandName, commandList...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
