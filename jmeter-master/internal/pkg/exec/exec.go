package exec

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tokuhirom/go-hsperfdata/hsperfdata"
	log "go.uber.org/zap"
)

//Exec use for executing bash scripts
type Exec struct {
	Env []string
}

var logger *log.Logger

func init() {
	logger, _ = log.NewProduction()
}

//IsProcessRunning checks to see if the process running
func (ex Exec) IsProcessRunning(process string) (running bool, pid string) {

	defer logger.Sync()
	os.Setenv("USER", "jmeter")
	repo, err := hsperfdata.New()
	if err == nil {
		files, err := repo.GetFiles()
		if err == nil {
			for _, f := range files {
				result, e := f.Read()
				if e == nil {
					name := result.GetProcName()
					fmt.Println(fmt.Sprintf("name=%s, pid=%s, process=%s", name, f.GetPid(), process))
					logger.Info("Running Process Details",
						// Structured context as strongly typed Field values.
						log.String("Name", name),
						log.String("Pid", f.GetPid()),
						log.String("Process", process),
						log.Duration("backoff", time.Second))
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
	defer logger.Sync()

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
							logger.Info("Running Process",
								// Structured context as strongly typed Field values.
								log.String("Name", name),
								log.String("Pid", f.GetPid()),
								log.Duration("backoff", time.Second))
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
