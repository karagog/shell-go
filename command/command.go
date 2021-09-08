// Package command implements an object for running shell commands.
package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Command is a shell command. Populate the member options and
// call Start() to run it. You can optionally consume the output by passing
// channels to Stdout/Stderr.
type Command struct {
	// Name is the command to execute.
	Name string

	// Args are the arguments to pass to the command.
	Args []string

	// Env contains environment variables in the form of FOO=BAR.
	// Tip: You can seed this with os.Environ() to inherit the environment
	// from the shell that called the program.
	Env []string

	// These channels, if non-nil, will be written to on each line of the output.
	// They must be fully consumed, and will be closed when the program is finished.
	Stdout chan<- string
	Stderr chan<- string

	// Stdin can be optionally given to provide input to the shell command.
	// If given, then it must be promptly closed when you're done pushing
	// data through it, because the command waits until stdin is closed.
	Stdin <-chan string
}

// Starts the command in a shell.
func (c *Command) Start(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Env = c.Env

	if c.Stdin != nil {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		go func() {
			for line := range c.Stdin {
				data := []byte(line)
				c, err := stdin.Write(data)
				if err != nil {
					panic(fmt.Errorf("cannot write stdin: %s", err))
				}
				if c != len(data) {
					panic(fmt.Errorf("wrote %d bytes, want %d", c, len(data)))
				}
			}

			// The writer closed the channel, so we go ahead and close the pipe.
			if err := stdin.Close(); err != nil {
				panic(err) // unable to close the stdin pipe
			}
		}()
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Reads the scanner until its output is finished, writing each line to the
	// channel. It closes the channel when done.
	read := func(s *bufio.Scanner, ch chan<- string, f *os.File) {
		if ch != nil {
			defer close(ch)
		}
		for s.Scan() {
			line := s.Text()
			if ch != nil {
				ch <- line
			} else {
				// Redirect the output by default if we're not given a channel.
				fmt.Fprintln(f, line)
			}
		}
	}

	go read(bufio.NewScanner(stdout), c.Stdout, os.Stdout)
	go read(bufio.NewScanner(stderr), c.Stderr, os.Stderr)
	return cmd, nil
}
