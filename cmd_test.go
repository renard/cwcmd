package cwcmd_test

import (
	"errors"
	"testing"

	"github.com/renard/cwcmd"
)

var (
	outArr []string
	errArr []string
)

func TestCmd(t *testing.T) {
	c := cwcmd.New(&cwcmd.Options{
		Buffered:  true,
		Streaming: true,
	}, "ls", "cmd_test.go")

	c.AddHook(processLines)

	if err := c.Start(); err != nil {
		t.Error(err)
	}

	if err := c.WaitStarted(); err != nil {
		t.Error(err)
	}

	i, err := c.Wait()
	if err != nil {
		t.Error(err)
	}

	if i != 0 {
		t.Error(errors.New("Return code is not 0."))
	}

	stdout := c.Cmd.Status().Stdout
	if !(len(stdout) > 0) || stdout[0] != "cmd_test.go" {
		t.Error(errors.New("Output should be 'cmd_test.go'."))
	}

	if len(errArr) > 0 {
		t.Error(errors.New("Error should be empty."))
	}

	if len(outArr) == 0 {
		t.Error(errors.New("Output stream should contain 1 line."))
	} else if outArr[0] != "cmd_test.go" {
		t.Error(errors.New("Output stream first line should contain 'cmd_test.go'."))
	}

}

func processLines(h *cwcmd.Hook) {
	defer close(h.Done)
	for h.Cmd.Stdout != nil || h.Cmd.Stderr != nil {
		select {
		case line, open := <-h.Cmd.Stdout:
			if !open {
				h.Cmd.Stdout = nil
				continue
			}
			outArr = append(outArr, line)
		case line, open := <-h.Cmd.Stderr:
			if !open {
				h.Cmd.Stderr = nil
				continue
			}
			errArr = append(errArr, line)
		}
	}
}
