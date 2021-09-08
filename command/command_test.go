package command

import (
	"context"
	"testing"
	"time"
)

func TestCommand_ConsumeOutput(t *testing.T) {
	const greeting = "Hello George"

	// Run a basic "echo" command to test stdoutCh.
	stdoutCh := make(chan string)
	c := Command{
		Name:   "echo",
		Args:   []string{greeting},
		Stdout: stdoutCh,
		// Leave Stderr nil for added coverage.
	}
	cmd, err := c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Consume the expected line.
	select {
	case got := <-stdoutCh:
		if got != greeting {
			t.Fatalf("Got output %q, want %q", got, greeting)
		}
	case <-time.After(time.Second):
		t.Fatal("No output received")
	}

	// The channel should be closed if we try to read it again.
	select {
	case _, ok := <-stdoutCh:
		if ok {
			t.Fatal("Stdout channel not closed")
		}
	case <-time.After(time.Second):
		t.Fatal("No output received")
	}

	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

// This test redirects the output instead of consuming it via channel.
func TestCommand_RedirectedOutput(t *testing.T) {
	// Run a basic "echo" command but don't consume the output to see if
	// it redirects to our output.
	c := Command{
		Name: "echo",
		Args: []string{"This output should be seen in our logs"},
	}
	cmd, err := c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

// This test pushes data to the shell command via stdin.
func TestCommand_Stdin(t *testing.T) {
	// Run a basic "echo" command but don't consume the output to see if
	// it redirects to our output.
	stdinCh := make(chan string)
	c := Command{
		// The grep command consumes stdin, and will write it to stdout because
		// we're matching on everything.
		Name:  "grep",
		Args:  []string{"."},
		Stdin: stdinCh,
	}
	cmd, err := c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Write a message to stdin.
	stdinCh <- "This output should be seen in our logs"

	// The program waits until stdin is closed, so we do that by closing
	// the channel.
	close(stdinCh)

	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}
