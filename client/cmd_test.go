package client

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func NewMockExec(stdout string, stderr string, rc int) func(command string, args ...string) *exec.Cmd {
	mock := func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"GO_HELPER_PROCESS_STDOUT=" + stdout,
			"GO_HELPER_PROCESS_STDERR=" + stderr,
			"GO_HELPER_PROCESS_RC=" + strconv.Itoa(rc),
		}
		return cmd
	}
	return mock
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	out := os.Getenv("GO_HELPER_PROCESS_STDOUT")
	if len(out) > 0 {
		fmt.Fprintf(os.Stdout, out)
	}
	out = os.Getenv("GO_HELPER_PROCESS_STDERR")
	if len(out) > 0 {
		fmt.Fprintf(os.Stderr, out)
	}
	out = os.Getenv("GO_HELPER_PROCESS_RC")
	if len(out) > 0 {
		i, _ := strconv.Atoi(out)
		os.Exit(i)
	}
	os.Exit(0)
}
