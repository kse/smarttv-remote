// +build linux

package main

import (
	"github.com/pkg/term/termios"
	"syscall"
)

func terminalRawInput(newmode *syscall.Termios) (current *syscall.Termios, e error) {
	current = &syscall.Termios{}
	e = termios.Tcgetattr(0, current)
	if e != nil {
		return
	}

	if newmode == nil {
		newmode = &syscall.Termios{}
		*newmode = *current

		lflag := ^(syscall.ICANON | syscall.ECHO)
		newmode.Lflag &= uint32(lflag)
	}

	termios.Tcsetattr(0, termios.TCSADRAIN, newmode)

	return current, nil
}
