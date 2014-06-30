package executil

import (
	"errors"
	"os/exec"
	"syscall"
)

type ExitStatus struct {
	ExitCode  int
	ExitError *exec.ExitError
}

func Run(cmd *exec.Cmd) (*ExitStatus, error) {
	err := cmd.Run()
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				return &ExitStatus{
					ExitCode:  s.ExitStatus(),
					ExitError: e2,
				}, nil
			} else {
				return &ExitStatus{
					ExitCode:  0,
					ExitError: e2,
				}, errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus.")
			}
		} else {
			return nil, err
		}
	}
	return &ExitStatus{
		ExitCode:  0,
		ExitError: nil,
	}, nil
}
