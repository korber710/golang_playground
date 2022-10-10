package run

import (
	"fmt"

	"github.com/korber710/golang_playground/robot-trigger/pkg/utility"
)

func Run(verbose bool, filename string) error {
	fmt.Println("Started Run()!", verbose)
	if verbose {
		return fmt.Errorf("Died")
	}

	// commands := []string{"./test/test1.robot"}
	commands := []string{filename}
	_ = utility.ExecShell("robot", commands)

	return nil
}
