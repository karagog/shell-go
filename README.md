# shell-go
Utilities for executing shell commands in Go.

Working with shell commands in Go can be a bit cumbersome for some use cases, so these utilities attempt to make life a bit easier. For example, you can use the Command class to run a command in a background goroutine and interact with standard in/out via channels.

## Basic Example - Run a command and redirect its output:

This example is somewhat contrived, because for this use case you'd probably be better off just using os/exec directly, but the example below shows how you can do it with this package as a baseline comparison.

```go
package main

import (
  "os"
  "github.com/karagog/shell-go/command"
)

func main() {
	c := command.Command{
		Name: "my_program",
		Args: []string{"arg1=foo"},
		Env:  append(os.Environ(), "BAR=BAZ"),
	}

	// Start the command running in the background.
	cmd, err := c.Start(context.Background())
	if err != nil {
		panic(err)
	}

	// Wait for the command to finish.
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
```

## Advanced Example - Write to stdin and consume stdout:

This is where the benefits of this package are more clear, because they allow us to interact with the running command using channels.

```go
package main

import "github.com/karagog/shell-go/command"

func main() {
	// Create channels to interact with the program's
	// standard input/output.
	stdinCh := make(chan string)
	stdoutCh := make(chan string)
	c := command.Command{
		Name: "my_program",
		Args: []string{"arg1=foo", "arg2=bar"},

		Stdin:  stdinCh,
		Stdout: stdoutCh,

		// Leaving any channel 'nil' is perfectly fine,
		// and means you don't care about it.
		Stderr: nil,
	}

	// Start the command running in the background.
	cmd, err := c.Start(context.Background())
	if err != nil {
		panic(err)
	}

	// Feed data on stdin. Make sure to close it when you're done,
	// otherwise the command will never end.
	stdinCh <- "Input Data"
	close(stdinCh)

	for line := range ch {
		// Process each line of the output as you like.
		fmt.Println("Got line: " + line)
	}

	// Wait for the command to finish.
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
```
