package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"text/template"

	"github.com/bbc/ugcuploader-test-kubernettes/admin/internal/pkg/kubernetes"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/magiconair/properties"
	uuid "github.com/satori/go.uuid"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// Code below taken from here: https://github.com/kjk/go-cookbook/blob/master/advanced-exec/03-live-progress-and-capture-v2.go

//FileUploadOperations used to perform file upload
type FileUploadOperations struct {
	Form   *multipart.Form
	Logger *log.Logger
}

//KubernetesOperations used for make calls to kubernetes
type KubernetesOperations struct {
	TestPath  string
	Tenant    string
	Bandwidth string
	Nodes     string
}

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

func (kop KubernetesOperations) startTest() (outStr string, errStr string) {

	args := []string{kop.TestPath, kop.Tenant, kop.Bandwidth, kop.Nodes}

	cmd := exec.Command("start_test_controller.sh", args...)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := NewCapturingPassThroughWriter(os.Stdout)
	stderr := NewCapturingPassThroughWriter(os.Stderr)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

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
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr = string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	return

}

func (kop KubernetesOperations) getGrafanaServiceHost() (outStr string, errStr string) {

	args := []string{"ugcload-reporter", "jmeter-grafana"}

	cmd := exec.Command("get-service-url.sh", args...)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := NewCapturingPassThroughWriter(os.Stdout)
	stderr := NewCapturingPassThroughWriter(os.Stderr)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

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
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr = string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	return
}

func (fop FileUploadOperations) processData() {

	dataFiles := fop.Form.File["data"]
	for _, f := range dataFiles {
		src, err := f.Open()
		if err != nil {
			fop.Logger.Fatalln("Unable to open file")
		}
		defer src.Close()

		// Destination
		data := props.MustGet("data")
		destFileName := fmt.Sprintf("%s/%s", data, f.Filename)
		fmt.Printf(fmt.Sprintf("destFileName=%s", destFileName))
		dst, err := os.Create(destFileName)
		fmt.Printf(fmt.Sprintf("m=%v", err))
		if err != nil {
			fop.Logger.Fatalln("Unable to create destination data file")
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			fop.Logger.Fatalln("Unable to copy source file to destination")
		}

	}

}

func (fop FileUploadOperations) processJmeter() (testFile string) {

	u2 := uuid.NewV4()
	dataFiles := fop.Form.File["jmeter"]
	path := fmt.Sprintf("%s/%s", props.MustGet("jmeter"), u2)
	fmt.Println(path)
	os.MkdirAll(path, os.ModePerm)
	for _, f := range dataFiles {
		src, err := f.Open()
		if err != nil {
			fop.Logger.Fatalln("Unable to open file")
		}
		defer src.Close()

		testFile = fmt.Sprintf("%s/%s", u2, f.Filename)
		destFileName := fmt.Sprintf("%s/%s", path, f.Filename)
		fmt.Printf(fmt.Sprintf("destFileName=%s", destFileName))
		dst, err := os.Create(destFileName)
		fmt.Printf(fmt.Sprintf("m=%v", err))
		if err != nil {
			fop.Logger.Fatalln("Unable to create destination data file")
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			fop.Logger.Fatalln("Unable to copy source file to destination")
		}

	}
	return
}

//UgcLoadRequest This is used to map to the form data.. seems to only work with firefox
type UgcLoadRequest struct {
	Context              string `json:"context" form:"context" validate:"required"`
	NumberOfNodes        string `json:"numberOfNodes" form:"numberOfNodes" validate:"numeric"`
	BandWidthSelection   string `json:"bandWidthSelection" numericform:"bandWidthSelection" validate:"required"`
	Jmeter               string `json:"jmeter" form:"jmeter" validate:"required"`
	Data                 string `json:"data" form:"data"`
	MissingTenant        bool
	MissingNumberOfNodes bool
	MissingJmeter        bool
	ProblemsBinding      bool
	MonitorUrl           string
	DashboardUrl         string
}

//CustomValidator based on: https://echo.labstack.com/guide/request
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func upload(c echo.Context) error {

	ugcLoadRequest := new(UgcLoadRequest)
	if err := c.Bind(ugcLoadRequest); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		return c.Render(http.StatusOK, "index.html", ugcLoadRequest)
	}

	if err := c.Validate(ugcLoadRequest); err != nil {
		for _, err := range err.(validator.ValidationErrors) {

			if err.Field() == "Context" {
				ugcLoadRequest.MissingTenant = true
			}
			if err.Field() == "NumberOfNodes" {
				ugcLoadRequest.MissingNumberOfNodes = true
			}
			if err.Field() == "Jmeter" {
				ugcLoadRequest.MissingJmeter = true
			}

		}
		return c.Render(http.StatusOK, "index.html", ugcLoadRequest)
	}

	fmt.Printf(fmt.Sprintf("a=%v b=%v", ugcLoadRequest.Context, ugcLoadRequest.NumberOfNodes))
	var (
		buf    bytes.Buffer
		logger = log.New(&buf, "logger: ", log.Lshortfile)
	)

	// Read form fields
	fmt.Printf("This is the tennand %v \n", ugcLoadRequest.Context)
	fmt.Printf("This is the noNodes %v \n", ugcLoadRequest.NumberOfNodes)
	fmt.Printf("This is the bandWidth %v \n", ugcLoadRequest.BandWidthSelection)
	fmt.Printf("This is the Jmeter %v \n", ugcLoadRequest.Jmeter)
	fmt.Printf("This is the Data %v \n", ugcLoadRequest.Data)

	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println("Unable to get the multi part form")
		return err
	}

	//fop := FileUploadOperations{Form: form, Logger: logger}
	_ = FileUploadOperations{Form: form, Logger: logger}
	//fop.processData()
	//testPath := fop.processJmeter()

	//fmt.Println(fmt.Sprintf("testPath=%s, tennat=%s, bandwidht=%s nonodes=%s", testPath, tenant, bandWidth, noNodes))
	//ko := KubernetesOperations{TestPath: testPath, Tenant: tenant, Bandwidth: bandWidth, Nodes: noNodes}
	//res, resError := ko.startTest()
	//grafanaHost, ghe := ko.getGrafanaServiceHost()
	//fmt.Println(fmt.Sprintf("grafanaHost=%s: errorFetchingGrafanHost=%s", grafanaHost, ghe))
	//return c.HTML(http.StatusOK, fmt.Sprintf("res=%s and resError=%s", res, resError))

	kubctlOps := kubernetes.KubernetesOperations{}
	kubctlOps.registerClient()
	kubctlOps.listPods()
	return c.Render(http.StatusOK, "index.html", ugcLoadRequest)
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Validator = &CustomValidator{validator: validator.New()}

	//ko := KubernetesOperations{}
	//response, err := ko.getGrafanaServiceHost()

	//temp := strings.Split(response, "\n")

	//grafanaHost := strings.Fields(temp[1])[0]
	//fmt.Println(fmt.Sprintf("grafanaHost=%s err=%s", grafanaHost, err))

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob(props.MustGet("web") + "/" + "*.html")),
	}
	e.Renderer = renderer

	formData := UgcLoadRequest{MonitorUrl: "http://monitor:8080", DashboardUrl: "http://dashboar"}
	// Named route "foobar"
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", formData)
	}).Name = "index"

	fmt.Println(fmt.Sprintf("%s/index.html", props.MustGet("web")))
	// Routes
	e.Static("/script", "web")
	e.POST("/start-test", upload)

	// Start server
	e.Logger.Debug(e.Start(":1323"))
}
