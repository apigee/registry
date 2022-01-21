package core

import (
	"os"
	"os/exec"
	"fmt"
	"testing"
)

type CommandGenerator func (string, ...string) *exec.Cmd

func GetCommandGenerator() CommandGenerator {
	return exec.Command
}

func GetFakeCommandGenerator(fakeProcessName string) CommandGenerator {
	fakeCommandGenerator := func(name string, args ...string) *exec.Cmd {
		fakeCommand := []string{fmt.Sprintf("-test.run=%s", fakeProcessName), "--", name}
		fakeCommand = append(fakeCommand, args...)
		cmd := exec.Command(os.Args[0], fakeCommand...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}

	return fakeCommandGenerator
}

func FakeTestProcess(t *testing.T, processArgs func(string, []string)) {
	t.Helper()

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	processArgs(cmd, args)
}