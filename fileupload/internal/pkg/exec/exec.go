package exec

import (
	"bytes"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Code below taken from here: https://github.com/kjk/go-cookbook/blob/master/advanced-exec/03-live-progress-and-capture-v2.go
// CapturingPassThroughWriter is a writer that remembers
// data written to it and passes it to w
type CapturingPassThroughWriter struct {
	buf bytes.Buffer
	w   io.Writer
}

// NewCapturingPassThroughWriter creates new CapturingPassThroughWriter
func NewCapturingPassThroughWriter(w io.Writer) *CapturingPassThroughWriter {
	return &CapturingPassThroughWriter{
		w: w,
	}
}

// Write writes data to the writer, returns number of bytes written and an error
func (w *CapturingPassThroughWriter) Write(d []byte) (int, error) {
	w.buf.Write(d)
	return w.w.Write(d)
}

// Bytes returns bytes written to the writer
func (w *CapturingPassThroughWriter) Bytes() []byte {
	return w.buf.Bytes()
}

//Exec use for executing bash scripts
type Exec struct{}

//ExecuteCommand used to execute the shell command
func (ex Exec) ExecuteCommand(command string, args []string) (outStr string, errStr string) {

	var logger = log.WithFields(log.Fields{
		"command": command,
		"args":    strings.Join(args, ","),
	})

	w := &logrusWriter{
		entry: logger,
	}

	cmd := exec.Command(command, args...)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := NewCapturingPassThroughWriter(w)
	stderr := NewCapturingPassThroughWriter(w)
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("unable to start the execute the command: %v", strings.Join(args, ","))
		errStr = err.Error()

	} else {

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			_, errStdout = io.Copy(stdout, stdoutIn)
			wg.Done()
		}()

		_, errStderr = io.Copy(stderr, stderrIn)
		wg.Wait()

		err = cmd.Wait()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Problems waiting for command to complete")
		}
		if errStdout != nil || errStderr != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Error occured when logging the execution process")
		}
		os, te := string(stdout.Bytes()), string(stderr.Bytes())

		if te != "" && strings.Contains(te, "TTY - input is not a terminal") {
			log.WithFields(log.Fields{
				"err": te,
			}).Warn("TTY - input is not a terminal: %v", strings.Join(args, ","))
		} else {
			errStr = te
		}
		outStr = os

	}
	return

}

type logrusWriter struct {
	entry *log.Entry
	buf   bytes.Buffer
	mu    sync.Mutex
}

func (w *logrusWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	origLen := len(b)
	for {
		if len(b) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			w.buf.Write(b)
			return origLen, nil
		}

		w.buf.Write(b[:i])
		w.alwaysFlush()
		b = b[i+1:]
	}
}

func (w *logrusWriter) alwaysFlush() {
	w.entry.Info(w.buf.String())
	w.buf.Reset()
}

func (w *logrusWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() != 0 {
		w.alwaysFlush()
	}
}
