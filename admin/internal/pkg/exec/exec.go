package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

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

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf, w)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf, w)

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
			errStr = err.Error()
			return
		}
		if errStdout != nil || errStderr != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("failed to capture stdout or stderr")
			errStr = err.Error()
			return
		}

		outStr, errStr = string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
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
