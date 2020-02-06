package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	myExec "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/exec"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/ugcupload"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
}

var (
	g errgroup.Group
)

//Upload used to save the data file
func Upload(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	fop.SaveFile(fmt.Sprintf("%s/%s", "/data", fileUpload.Name))
}

//JmeterProps used to save the jmeter
func JmeterProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/jmeter.properties", jh))
}

func checkFileUploadLogs() (found bool) {
	file, err := os.Open("/fileupload.log")

	if err == nil {
		log.Fatalf("failed opening file: %s", err)

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		var txtlines []string

		for scanner.Scan() {
			if strings.Contains(scanner.Text().ToLower(), "Caused by: java.net.ConnectException: Connection refused (Connection refused)")
		}

		file.Close()

		for _, eachline := range txtlines {
			fmt.Println(eachline)
		}
	}

	found = false
	return
}

//IsRunning Used to determing if slave is running
func IsRunning(c *gin.Context) {
	cmd := myExec.Exec{}
	running := "no"
	if cmd.IsProcessRunning("ApacheJMeter.jar") {
		running = "yes"
	}
	c.String(http.StatusOK, running)
	return
}

//UserProps user to save the user.properties
func UserProps(c *gin.Context) {
	fileUpload := new(FileUpload)
	fop := ugl.FileUploadOperations{Context: c}
	if err := c.ShouldBindWith(fileUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form")
		return
	}
	jh := os.Getenv("JMETER_HOME")
	fop.SaveFile(fmt.Sprintf("%s/bin/user.properties", jh))
}

func start(cmd string, args ...string) (p *os.Process, err error) {
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin,
		os.Stdout, os.Stderr}
	p, err = os.StartProcess(cmd, args, &procAttr)
	return
}

//StartJmeterServer NOTE: Had to do this because the bash script was hanging...
func startJmeterServer(startUpload StartUpload) (started bool) {

	jvmArgs := []string{fmt.Sprintf("JVM_ARGS=%s", fmt.Sprintf("-Xms%sg -Xmx%sg -XX:MaxMetaspaceSize=%sm",
		startUpload.Xms, startUpload.Xmx, startUpload.MaxMetaspaceSize))}

	cmd := myExec.Exec{Env: jvmArgs}
	start, pid := cmd.ExecuteCommandSlaveCommand("/start.sh", []string{})
	if start {
		go func() {
			cmd = myExec.Exec{}
			args := []string{"-jar", "/fileupload/jolokia-jvm-1.6.2-agent.jar", "start", pid}
			cmd = myExec.Exec{}
			_, _ = cmd.ExecuteCommand("java", args)
		}()
		started = true
	}
	return
}

//StartServer used to start jmeter server
func StartServer(c *gin.Context) {

	startUpload := new(StartUpload)
	if err := c.ShouldBindWith(startUpload, binding.Form); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems binding form to start the upload")
		c.String(http.StatusBadRequest, "Unable to start the test - problems binding form")
		return
	}

	started := startJmeterServer(*startUpload)
	//Just giving jmeter server time to start
	time.Sleep(2 * time.Second)
	if started {
		c.String(http.StatusOK, "ok")
		return
	} else {
		c.String(http.StatusBadRequest, "no")
	}
}

func main() {

	server01 := &http.Server{
		Addr:         ":1007",
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

//FileUpload This is the struct returned
type FileUpload struct {
	File string `json:"file" form:"file"`
	Name string `json:"name" form:"name"`
}

//StartUpload holds the values for configuring the slave
type StartUpload struct {
	Xmx              string `json:"xmx" form:"xmx"`
	Xms              string `json:"xms" form:"xms"`
	MaxMetaspaceSize string `json:"maxMetaspaceSize" form:"maxMetaspaceSize"`
}

func router01() http.Handler {
	// Gin instance
	r := gin.Default()

	gob.Register(FileUpload{})
	gob.Register(StartUpload{})

	r.POST("/data", Upload)
	r.POST("/jmeter-props", JmeterProps)
	r.POST("/user-propes", UserProps)
	r.POST("/start-server", StartServer)
	r.GET("/is-running", IsRunning)

	return r
}
