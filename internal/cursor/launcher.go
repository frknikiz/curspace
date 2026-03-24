package cursor

import (
	"fmt"
	"os/exec"
	"runtime"
)

func Open(workspacePath string) error {
	if path, err := exec.LookPath("cursor"); err == nil {
		cmd := exec.Command(path, workspacePath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting cursor: %w", err)
		}
		return nil
	}

	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "Cursor", workspacePath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("opening Cursor via macOS open command: %w", err)
		}
		return nil
	}

	return fmt.Errorf("cursor command not found. Please install Cursor and ensure 'cursor' is in your PATH")
}
