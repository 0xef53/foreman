package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"

	"github.com/nsqio/go-nsq"
)

// byLineCopy reads data from the pipe, prepends every line with a given prefix string
// and writes results to the Writer (e.g. stdout/stderr/file).
func byLineCopy(prefix string, sink io.Writer, pipe io.Reader) {
	wg.Add(1)
	defer wg.Done()

	prefix += " "
	buf := []byte(prefix)

	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		buf := buf[:len(prefix)]
		buf = append(buf, scanner.Bytes()...)
		buf = append(buf, '\n')
		// this is safe to write a single line to stdout/stderr without
		// additional locking from multiple goroutines as OS guarantees
		// those writes are atomic (for stdout/stderr only)
		sink.Write(buf)
	}
	if err := scanner.Err(); err != nil {
		switch err.(type) {
		case *os.PathError:
		default:
			Logger.Println("string scanner error:", err)
		}
	}
}

// executeCommand takes a command template, a payload from the NSQ message,
// merges them and trying to execute a resulting command.
//
// A payload data is also passed to the stdin of the executing command.
func executeCommand(cmdTpl, prefix string, data map[string]interface{}, msg *nsq.Message, envs []string) (int, error) {
	t := template.Must(template.New("cmd").Parse(cmdTpl))
	var cmdB bytes.Buffer
	if err := t.Execute(&cmdB, data); err != nil {
		return -1, err
	}

	cmdArgs := strings.Fields(cmdB.String())

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	cmd.Stdin = bytes.NewReader(append(msg.Body, '\n'))
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return -1, fmt.Errorf("cannot create stdout pipe: %s", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return -1, fmt.Errorf("cannot create stderr pipe: %s", err)
	}

	go byLineCopy(prefix, os.Stdout, stdoutPipe)
	go byLineCopy(prefix, os.Stderr, stderrPipe)

	cmd.Env = append(os.Environ(), envs...)

	if err := cmd.Start(); err != nil {
		return -1, err
	}

	if err := cmd.Wait(); err != nil {
		var exitCode int
		if exiterr, ok := err.(*exec.ExitError); ok {
			status := exiterr.Sys().(syscall.WaitStatus)

			switch {
			case status.Exited():
				exitCode = status.ExitStatus()
			case status.Signaled():
				exitCode = 128 + int(status.Signal())
			}

			if exitCode == 100 {
				return exitCode, NewRequeueError()
			}
		}
		return exitCode, err
	}

	return 0, nil
}

// RequeueError means that the worker command should be restarted
// if it allows the maximum number of attempts specified for this topic.
type RequeueError struct {
}

// NewRequeueError returns a new instance of error.
func NewRequeueError() error {
	return &RequeueError{}
}

// Error returns a string representation of error.
func (e *RequeueError) Error() string {
	return fmt.Sprintf("re-queue needed")
}

// IsRequeueError returns a boolean indicating whether the type of error is RequeueError.
func IsRequeueError(err error) bool {
	if _, ok := err.(*RequeueError); ok {
		return true
	}
	return false
}
