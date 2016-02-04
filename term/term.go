package term

import (
	"os"
	"os/exec"
	"runtime"
)

func Clear() {
	var cmd *exec.Cmd
	switch os := runtime.GOOS; os {
	case "linux":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
