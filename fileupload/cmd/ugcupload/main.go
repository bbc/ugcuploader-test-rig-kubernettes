package main

import (
	"archive/zip"
	"bufio"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	myExec "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/exec"
	testlogs "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/testlogs"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/fileupload/internal/pkg/ugcupload"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type LogFileState int ////not visible outside of the package unary

const (
	NotExist LogFileState = iota
	ConnectionRefused
	TestStop
	Running
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

func checkFileUploadLogs() (logFileState LogFileState) {
	file, err := os.Open("/fileupload.log")

	if err != nil {
		logFileState = NotExist
		return
	}
	if err == nil {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower("Caused by: java.net.ConnectException: Connection refused (Connection refused)")) {
				logFileState = ConnectionRefused
				return
			}
			if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower("stop")) {
				logFileState = TestStop
				return
			}

			if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower("JMeterThread: Thread started")) {
				logFileState = Running
				return
			}

		}
		file.Close()
	}

	logFileState = Running
	return
}

//Kill Used to stop slaves immediately
func Kill(c *gin.Context) {
	cmd := myExec.Exec{}
	found, pidStr := cmd.IsProcessRunning("ApacheJMeter.jar")
	if found {
		pid, _ := strconv.Atoi(pidStr)
		syscall.Kill(pid, 9)
		if err := os.RemoveAll("/tmp/hsperfdata_jmeter"); err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Removing hsperfdata_jmeter")
		}
	}
}

//IsRunning Used to determine if slave is running
func IsRunning(c *gin.Context) {
	cmd := myExec.Exec{}
	state := "notstarted"
	if found, _ := cmd.IsProcessRunning("ApacheJMeter.jar"); found {

		switch checkFileUploadLogs() {
		case TestStop:
			state = "teststop"
		case NotExist:
			state = "notexit"
		case ConnectionRefused:
			state = "connectionrefused"
		default:
			state = "running"
		}
		state = "yes"
	}
	c.String(http.StatusOK, state)
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

//TestOutput used to fetch the test log file from the server
func TestOutput(c *gin.Context) {
	hn := os.Getenv("HOSTNAME")
	c.Writer.Header().Set("Content-type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename='%s.zip'", hn))
	ar := zip.NewWriter(c.Writer)
	tl := testlogs.ZipTestoutput{ZipWriter: ar}
	tl.ZipLogFiles()
	ar.Close()
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
	r.GET("/kill", Kill)
	r.GET("/test-output", TestOutput)

	return r
}
