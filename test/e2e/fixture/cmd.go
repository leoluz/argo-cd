package fixture

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	argoexec "github.com/argoproj/pkg/exec"
	"github.com/sirupsen/logrus"
)

func Run(workDir, name string, args ...string) (string, error) {
	return RunWithStdin("", workDir, name, args...)
}

func RunWithStdin(stdin, workDir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	cmd.Env = os.Environ()
	cmd.Dir = workDir

	return argoexec.RunCommandExt(cmd, argoexec.CmdOpts{})
}

// RunProcess will execute the given command in a goroutine and block until
// the given context is canceled. This is useful for e2e tests that need to
// start a process to be running during the execution. The test should cancel
// the context once done. The process will timeout after 60 seconds.
func RunProcess(ctx context.Context, workDir, command string, env map[string]string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Dir = workDir
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output := ""
	var err error
	go func(cmd *exec.Cmd) {
		output, err = argoexec.RunCommandExt(cmd, argoexec.CmdOpts{})
	}(cmd)

	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			// cmd.Process.Signal(syscall.SIGTERM)
			cmd.Process.Kill()
		}
	case <-time.After(60 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
	if err != nil {
		logrus.Errorf("error executing process command: %s output: %s", err, output)
	}
}
