// Copyright © 2020 Sébastien Gross <seb•ɑƬ•chezwam•ɖɵʈ•org>
//
// Created: 2020-02-28
// Last changed: 2020-02-28 02:45:23
//
// This program is free software. It comes without any warranty, to
// the extent permitted by applicable law. You can redistribute it
// and/or modify it under the terms of the Do What The Fuck You Want
// To Public License, Version 2, as published by Sam Hocevar. See
// http://sam.zoy.org/wtfpl/COPYING for more details.

// Package cwcmd is a simple wrapper around go-cmd/cmd package. Its main
// purpose is to simplify its usage (if this is possible) and add a generic
// purpose hook.
//
// The typical usage of this hook is to perform all logging operations and
// stream the command output to the terminal.
//
//
//
//   func main() {
//   	c := cwcmd.New(&cwcmd.Options{
//   		Buffered:  false,
//   		Streaming: true,
//   	}, "ls", "-l", "/")
//
//   	c.AddHook(PrintLines)
//   	if err := c.Start(); err != nil {
//   		panic(err)
//   	}
//
//   	i, err := c.Wait()
//   	fmt.Printf("Done: %d\n", i)
//   	if err != nil {
//   		panic(err)
//   	}
//   }
//
//   func PrintLines(h *cwcmd.Hook) {
//   	fmt.Printf("Starting command: %s %s\n", h.Cmd.Name, h.Cmd.Args)
//   	defer close(h.Done)
//   	for h.Cmd.Stdout != nil || h.Cmd.Stderr != nil {
//   		select {
//   		case line, open := <-h.Cmd.Stdout:
//   			if !open {
//   				h.Cmd.Stdout = nil
//   				continue
//   			}
//   			fmt.Println(line)
//   		case line, open := <-h.Cmd.Stderr:
//   			if !open {
//   				h.Cmd.Stderr = nil
//   				continue
//   			}
//   			fmt.Fprintln(os.Stderr, line)
//   		}
//   	}
//   	fmt.Printf("Exit: %d\n", h.Cmd.Status().Exit)
//   }
package cwcmd

import (
	"time"

	"github.com/go-cmd/cmd"
)

// Cmd is the global structure wrapping cmd.Cmd struct.
type Cmd struct {
	// Cmd is the cmd.Cmd struct. See its documentation for this one. All
	// cmd.Cmd methods are accessible from here.
	*cmd.Cmd
	// A Hook struct added by AddHook.
	hook *Hook
	// A channel used by Wait to detect the process termination.
	chanCmd <-chan cmd.Status
}

// Options controls the cmd.Cmd behavior. This is a convenient struct
// preventing from importing github.com/go-cmd/cmd all the time.
//
// See cmd.Options documentation.
type Options struct {
	// Capture output into Cmd buffers. Use this if output should be
	// processed.
	Buffered bool
	// Stream output into channel. Use this to display output on terminal.
	Streaming bool
}

type hookFunc func(*Hook)
type hookChan struct{}

// Hook struct passed as parameter to the hook function.
type Hook struct {
	*cmd.Cmd
	// A channel that the hook must close when it finishes to prevent
	// execution from hanging.
	Done chan hookChan
	f    hookFunc
}

// New returns an Cmd struct ready to run command and its optional args.
// Use Options to capture output.
func New(o *Options, command string, args ...string) (c *Cmd) {
	c = &Cmd{}
	c.Cmd = cmd.NewCmdOptions(cmd.Options{
		Buffered:  o.Buffered,
		Streaming: o.Streaming,
	}, command, args...)
	return c
}

// AddHook add a new Hook to Cmd. When Cmd is started the function f is run
// as a gorouting.
func (c *Cmd) AddHook(f hookFunc) {
	c.hook = c.newHook(f)
}

// Wait waits for the command to terminate. If a hook is defined, it also
// wait for the hook function to terminate. It is mandatory for the hook to
// close the Done channel to prevent from hanging.
//
// The command exit code is returned along the status error.
func (c *Cmd) Wait() (int, error) {
	<-c.chanCmd
	if c.hook != nil {
		<-c.hook.Done
	}
	return c.Cmd.Status().Exit, c.Cmd.Status().Error
}

// WaitStarted waits until the command has started. Use this method to run a
// process in backgound. If main programm exit too fast, command may not
// have been started.
//
// In most of case both buffered and streaming option should be disabled for
// this functionnality to work. Note that main program should also exit in a
// clean way.
//
// The Cmd.Status().Error is returned for error checking.
func (c *Cmd) WaitStarted() error {
	for c.Cmd.Status().StartTs <= 0 {
		// Do not emulate CPU burn. It takes less than 5ms to start a
		// command.
		time.Sleep(1 * time.Millisecond)
	}
	return c.Cmd.Status().Error
}

// Start starts the cmd.Cmd and returns immediately. If the command must
// live after main program exited the WaitStarted function must be
// called. Is the result of the command is required to continue the program
// execution, the Wait function must be called.
//
// If a hook is defined it is started in a goroutine before the command
// really starts.
//
// The Cmd.Status().Error is returned for error checking.
func (c *Cmd) Start() error {
	if c.hook != nil {
		go c.hook.f(c.hook)
	}
	c.chanCmd = c.Cmd.Start()
	return c.Cmd.Status().Error
}

// Creates a new Hook using f function. It also create the Done hookChan.
func (c *Cmd) newHook(f hookFunc) *Hook {
	h := &Hook{
		Cmd:  c.Cmd,
		f:    f,
		Done: make(chan hookChan),
	}
	return h
}
