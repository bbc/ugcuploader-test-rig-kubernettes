package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tokuhirom/go-hsperfdata/hsperfdata"
)

//Exec use for executing bash scripts
type Exec struct {
	Env []string
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

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

//IsProcessRunning checks to see if the process running
func (ex Exec) IsProcessRunning(process string) (running bool, pid string) {
	os.Setenv("USER", "root")
	repo, err := hsperfdata.New()
	if err == nil {
		files, err := repo.GetFiles()
		if err == nil {
			for _, f := range files {
				result, e := f.Read()
				if e == nil {
					name := result.GetProcName()
					log.WithFields(log.Fields{
						"Name":    name,
						"Pid":     f.GetPid(),
						"Process": process,
					}).Info("Running Process")
					if strings.Contains(strings.ToLower(name), strings.ToLower(process)) {
						running = true
						pid = f.GetPid()
						return
					}
				}
			}
		}
	}
	return
}

//WaitForProcessToStart waiting for the process to start
func (ex Exec) WaitForProcessToStart(process string) (started bool, pid string) {

	os.Setenv("USER", "root")
	count := 0
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for {
		select {
		case <-ctx.Done():

			repo, err := hsperfdata.New()
			if err == nil {
				files, err := repo.GetFiles()
				if err == nil {
					for _, f := range files {
						result, e := f.Read()
						if e == nil {
							name := result.GetProcName()
							log.WithFields(log.Fields{
								"Name": name,
								"Pid":  f.GetPid(),
							}).Info("Running Process")
							if strings.Contains(strings.ToLower(name), strings.ToLower(process)) {
								started = true
								pid = f.GetPid()
								return
							}
						}
					}
				}
			}
			cancel()
			count = count + 1
			//Waited too long for slave to start..something went wrong
			if count == 4 {
				started = false
				return
			}
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		}
	}
}

//ExecuteCommandSlaveCommand used to start the slave
func (ex Exec) ExecuteCommandSlaveCommand(command string, args []string) (started bool, pid string) {

	started = false
	cmd := exec.Command(command, args...)

	cmd.Env = append(os.Environ(), ex.Env...)
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("unable to start the execute the command: %v", strings.Join(args, ","))

	} else {
		go func() {
			err = cmd.Wait()
			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Info("Slave has ended with an error")
			}
		}()
		return ex.WaitForProcessToStart("ApacheJMeter.jar")
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
