package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

//Allows tests to mock this command
var execCommand = exec.Command

func errorIndent(content string) string {
	return "| " + strings.Replace(content, "\n", "\n| ", -1) + "_"
}

func RunFrom(fromDir string, command string, args ...string) (string, error) {
	cmd := execCommand(command, args...)
	cmd.Dir = fromDir
	binaryOut, err := cmd.CombinedOutput()
	out := string(binaryOut)
	if err != nil {
		return "", fmt.Errorf("Unable to run '%s'. err(%s), output=\n%s",
			cmd.Args, err, errorIndent(out))
	}

	return out, nil
}

func Run(command string, args ...string) (string, error) {
	return RunFrom("", command, args...)
}

func RunFromStreamedTo(fromDir string, stdOut, stdErr io.Writer, command string, args ...string) error {
	cmd := execCommand(command, args...)
	cmd.Dir = fromDir
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Unable to run '%s': err=%s", cmd.Args, err)
	}
	return nil
}

func RunFromStreamed(fromDir string, command string, args ...string) error {
	return RunFromStreamedTo(fromDir, os.Stdout, os.Stderr, command, args...)
}

func RunStreamed(command string, args ...string) error {
	return RunFromStreamedTo("", os.Stdout, os.Stderr, command, args...)
}
