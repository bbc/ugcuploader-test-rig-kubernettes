package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	cluster "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/cluster"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"

	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/magiconair/properties"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

var nonValidNamespaces = []string{"control", "default", "kube-node-lease", "kube-public", "kube-system", "ugcload-reporter", "weave"}

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)
var kubctlOps = kubernetes.Operations{}

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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//UgcLoadRequest This is used to map to the form data.. seems to only work with firefox
type UgcLoadRequest struct {
	Context              string `json:"context" form:"context" validate:"required"`
	NumberOfNodes        int    `json:"numberOfNodes" form:"numberOfNodes" validate:"numeric,min=1"`
	BandWidthSelection   string `json:"bandWidthSelection" numericform:"bandWidthSelection" validate:"required"`
	Jmeter               string `json:"jmeter" form:"jmeter"`
	Data                 string `json:"data" form:"data"`
	MissingTenant        bool
	MissingNumberOfNodes bool
	MissingJmeter        bool
	ProblemsBinding      bool
	MonitorURL           string
	DashboardURL         string
	Success              string
	InvalidTenantName    string
	TenantDeleted        string
	TenantContext        string
	TenantMissing        bool
	InvalidTenantDelete  string
	TennantNotDeleted    string
	GenericCreateTestMsg string
	StopContext          string
	StopTenantMissing    bool
	InvalidTenantStop    string
	TennantNotStopped    string
	TenantStopped        string
}

//CustomValidator based on: https://echo.labstack.com/guide/request
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func addMonitorAndDashboard(ur *UgcLoadRequest) {
	ur.MonitorURL = fmt.Sprintf("http://%s:4040", kubctlOps.LoadBalancerIP("control"))
	ur.DashboardURL = fmt.Sprintf("http://%s:3000", kubctlOps.LoadBalancerIP("ugcload-reporter"))
}

func stopTest(c echo.Context) error {
	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)
	addMonitorAndDashboard(ugcLoadRequest)
	if err := c.Bind(ugcLoadRequest); err != nil {
		addMonitorAndDashboard(ugcLoadRequest)
		ugcLoadRequest.ProblemsBinding = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}
	log.WithFields(log.Fields{
		"StopContext": ugcLoadRequest.StopContext,
	}).Info("StopContext")

	if ugcLoadRequest.StopContext == "" {
		ugcLoadRequest.StopTenantMissing = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	if stringInSlice(ugcLoadRequest.StopContext, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantStop = strings.Join(nonValidNamespaces, ",")
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}
	deleted, errStr := kubctlOps.StopTest(ugcLoadRequest.StopContext)
	if deleted == false {
		log.WithFields(log.Fields{
			"Context": ugcLoadRequest.StopContext,
			"err":     errStr,
		}).Info("Unable to stop the test")
		ugcLoadRequest.TennantNotStopped = fmt.Sprintf("Unable to stop tenant: %s", errStr)
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	ugcLoadRequest.TenantStopped = ugcLoadRequest.StopContext
	ugcLoadRequest.StopContext = ""
	return c.Render(http.StatusOK, "index.html", ugcLoadRequest)
}

func deleteTenant(c echo.Context) error {

	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)
	addMonitorAndDashboard(ugcLoadRequest)
	if err := c.Bind(ugcLoadRequest); err != nil {
		addMonitorAndDashboard(ugcLoadRequest)
		ugcLoadRequest.ProblemsBinding = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	log.WithFields(log.Fields{
		"TenantContext": ugcLoadRequest.TenantContext,
	}).Info("TenantContext")

	if ugcLoadRequest.TenantContext == "" {
		ugcLoadRequest.TenantMissing = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	if stringInSlice(ugcLoadRequest.TenantContext, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantDelete = strings.Join(nonValidNamespaces, ",")
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	deleted, errStr := kubctlOps.DeleteServiceAccount(ugcLoadRequest.TenantContext)
	log.WithFields(log.Fields{
		"deleted": deleted,
	}).Info("Deleted")

	if deleted == false {
		log.WithFields(log.Fields{
			"Context": ugcLoadRequest.TenantContext,
			"err":     errStr,
		}).Info("UnableToDeleteServiceAccount")
		ugcLoadRequest.TennantNotDeleted = fmt.Sprintf("Unable to delete service account: %s", errStr)
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}
	ugcLoadRequest.TenantDeleted = ugcLoadRequest.TenantContext
	ugcLoadRequest.TenantContext = ""
	return c.Render(http.StatusOK, "index.html", ugcLoadRequest)

}

func upload(c echo.Context) error {

	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)
	ugcLoadRequest.MonitorURL = fmt.Sprintf("http://%s:4040", kubctlOps.LoadBalancerIP("control"))
	ugcLoadRequest.DashboardURL = fmt.Sprintf("http://%s:3000", kubctlOps.LoadBalancerIP("ugcload-reporter"))
	if err := c.Bind(ugcLoadRequest); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
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
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	if stringInSlice(ugcLoadRequest.Context, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantName = strings.Join(nonValidNamespaces, ",")
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	form, err := c.MultipartForm()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Info("Request Properties")
		kubctlOps.RegisterClient()
		ugcLoadRequest.MissingJmeter = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	fop := ugl.FileUploadOperations{Form: form}
	testPath := fop.ProcessJmeter()
	if testPath == "" {
		_ = fop.ProcessData()
		kubctlOps.RegisterClient()
		ugcLoadRequest.MissingJmeter = true
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	nsExist := kubctlOps.CheckNamespaces(ugcLoadRequest.Context)
	if nsExist == false {
		clusterops := cluster.Operations{}
		awsRegion, awsAcntNumber := clusterops.DescribeCluster("ugcloadtest")
		log.WithFields(log.Fields{
			"awsAcntNumber": awsAcntNumber,
			"awsRegion":     awsRegion,
		}).Info("Cluster Info")

		aan, err := strconv.ParseInt(awsAcntNumber, 10, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Unable to Parse Integer")
			ugcLoadRequest.GenericCreateTestMsg = err.Error()
			return c.Render(http.StatusConflict, "index.html", ugcLoadRequest)
		}
		created, errNs := kubctlOps.CreateNamespace(ugcLoadRequest.Context)
		if created == false {
			log.WithFields(log.Fields{
				"err": errNs,
			}).Error("Unable to create namespace")
			ugcLoadRequest.GenericCreateTestMsg = errNs
			return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
		}
		policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/ugcupload-eks-jmeter-policy", awsAcntNumber)
		crtd, e := kubctlOps.CreateServiceaccount(ugcLoadRequest.Context, policyArn)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create ServiceAccount")
			ugcLoadRequest.GenericCreateTestMsg = e
			return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
		}
		crtd, e = kubctlOps.CreateJmeterMasterDeployment(ugcLoadRequest.Context, aan, awsRegion)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Master Deployment")
			ugcLoadRequest.GenericCreateTestMsg = e
			return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
		}
		crtd, e = kubctlOps.CreateJmeterSlaveService(ugcLoadRequest.Context)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Slave Service")
			ugcLoadRequest.GenericCreateTestMsg = e
			return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
		}
		crtd, e = kubctlOps.CreateJmeterSlaveDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes), aan, awsRegion)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Slave Deployment")
			ugcLoadRequest.GenericCreateTestMsg = e
			return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
		}

	}

	log.WithFields(log.Fields{
		"ugcLoadRequest.Context":            ugcLoadRequest.Context,
		"ugcLoadRequest.NumberOfNodes":      ugcLoadRequest.NumberOfNodes,
		"ugcLoadRequest.BandWidthSelection": ugcLoadRequest.BandWidthSelection,
		"ugcLoadRequest.Jmeter":             ugcLoadRequest.Jmeter,
		"ugcLoadRequest.Data":               ugcLoadRequest.Data,
	}).Info("Request Properties")

	started, errStartTest := kubctlOps.StartTest(testPath, ugcLoadRequest.Context, ugcLoadRequest.BandWidthSelection, ugcLoadRequest.NumberOfNodes)
	if started == false {
		ugcLoadRequest.GenericCreateTestMsg = errStartTest
		return c.Render(http.StatusBadRequest, "index.html", ugcLoadRequest)
	}

	ugcLoadRequest.Success = fmt.Sprintf("Test %s was succesfully created for tenant[%s]", testPath, ugcLoadRequest.Context)
	ugcLoadRequest.NumberOfNodes = 0
	ugcLoadRequest.Context = ""
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

	kubctlOps.Init()
	kubctlOps.RegisterClient()
	formData := UgcLoadRequest{MonitorURL: fmt.Sprintf("http://%s:4040", kubctlOps.LoadBalancerIP("control")),
		DashboardURL: fmt.Sprintf("http://%s:3000", kubctlOps.LoadBalancerIP("ugcload-reporter"))}

	// Named route "foobar"
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", formData)
	}).Name = "index"

	// Routes
	e.Static("/script", "web")
	e.POST("/start-test", upload)
	e.POST("/stop-test", stopTest)
	e.POST("/delete-tenant", deleteTenant)

	s := &http.Server{
		Addr:         ":1323",
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}

	//Taken from here: https://stackoverflow.com/questions/29334407/creating-an-idle-timeout-in-go/29334926#29334926
	s.ConnState = func(c net.Conn, cs http.ConnState) {
		switch cs {
		case http.StateIdle, http.StateNew:
			c.SetReadDeadline(time.Now().Add(time.Minute * 5))
		case http.StateActive:
			c.SetReadDeadline(time.Now().Add(time.Minute * 5))
		}
	}

	// Start server
	e.Logger.Debug(e.StartServer(s))
}
