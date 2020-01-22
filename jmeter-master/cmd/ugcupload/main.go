package main

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"

	shellExec "github.com/bbc/ugcuploader-test-rig-kubernettes/jmeter-master/internal/pkg/exec"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

}

var (
	g errgroup.Group
)

//StopTest used to stop the tests
func StopTest(c *gin.Context) {

	cmd := fmt.Sprintf("/opt/apache-jmeter/bin/stoptest.sh")

	args := []string{"<", "/dev/null"}
	se := shellExec.Exec{}
	_, err := se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start the test")
	}
	return
}

func main() {

	server01 := &http.Server{
		Addr:         ":1008",
		Handler:      router01(),
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	g.Go(func() error {
		return server01.ListenAndServe()
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

}

func router01() http.Handler {
	// Gin instance
	r := gin.Default()

	r.GET("/stop-test", StopTest)

	return r
}
