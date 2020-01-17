package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	aws "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/aws"
	cluster "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/cluster"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"

	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin/binding"
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

// Gin instance
var r = gin.Default()

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
	TenantContext        string `json:"TenantContext" form:"TenantContext"`
	TenantMissing        bool
	InvalidTenantDelete  string
	TennantNotDeleted    string
	GenericCreateTestMsg string
	StopContext          string `json:"stopcontext" form:"stopcontext"`
	StopTenantMissing    bool
	InvalidTenantStop    string
	TennantNotStopped    string
	TenantStopped        string
	TenantList           []string
	ReportURL            string
}

//CustomValidator based on: https://echo.labstack.com/guide/request
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func addMonitorAndDashboard(ur *UgcLoadRequest) {
	ip := kubctlOps.LoadBalancerIP("control")
	ur.MonitorURL = fmt.Sprintf("http://%s:4040", ip)
	ur.ReportURL = fmt.Sprintf("http://%s:80", ip)
	ur.DashboardURL = fmt.Sprintf("http://%s:3000", kubctlOps.LoadBalancerIP("ugcload-reporter"))
}

func addTenants(ur *UgcLoadRequest) {
	s3Ops := aws.S3Operations{}
	ur.TenantList, _ = s3Ops.GetBucketItems("ugcupload-jmeter", "", 0)
}

func allTenants(c *gin.Context) {
	kubctlOps.RegisterClient()
	t, e := kubctlOps.GetallTenants()

	if e != "" {
		c.String(http.StatusBadRequest, "Unable to fetch tenants")
	}

	nt := []kubernetes.Tenant{}

	for _, tenant := range t {
		r, e := kubctlOps.CheckIfRunningJava(tenant.Namespace, tenant.Name)
		if len(e) > 0 || len(r) < 1 {
			tenant.Running = false
		} else {
			tenant.Running = true
		}
		nt = append(nt, tenant)
	}
	r, _ := json.Marshal(nt)
	c.String(http.StatusOK, string(r))
}

func generateReport(c *gin.Context) {

	tenant, _ := c.GetPostForm("tenant")
	data, _ := c.GetPostForm("data")

	var items []string
	for _, d := range strings.Split(data, ",") {
		items = append(items, fmt.Sprintf("%s=%s", tenant, d))
	}
	kubctlOps.RegisterClient()
	_, e := kubctlOps.GenerateReport(strings.Join(items[:], ","))
	c.String(http.StatusOK, e)
}

func s3Tenants(c *gin.Context) {

	type Items struct {
		Date string `json:"date"`
	}
	s3Ops := aws.S3Operations{}
	tenant, _ := c.GetQuery("tenant")

	var my []Items
	items, _ := s3Ops.GetBucketItems("ugcupload-jmeter", fmt.Sprintf("%s/", tenant), 1)
	for _, item := range items {
		it := Items{Date: item}
		my = append(my, it)
	}
	c.JSON(http.StatusOK, &my)
	return
}

func stopTest(c *gin.Context) {

	session := sessions.Default(c)
	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)

	if err := c.ShouldBindWith(&ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	log.WithFields(log.Fields{
		"StopContext": ugcLoadRequest.StopContext,
	}).Info("StopContext")

	if ugcLoadRequest.StopContext == "" {
		ugcLoadRequest.StopTenantMissing = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	if stringInSlice(ugcLoadRequest.StopContext, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantStop = strings.Join(nonValidNamespaces, ",")
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	deleted, errStr := kubctlOps.StopTest(ugcLoadRequest.StopContext)
	if deleted == false {
		log.WithFields(log.Fields{
			"Context": ugcLoadRequest.StopContext,
			"err":     errStr,
		}).Info("Unable to stop the test")
		ugcLoadRequest.TennantNotStopped = fmt.Sprintf("Unable to stop tenant: %s", errStr)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	ugcLoadRequest.TenantStopped = ugcLoadRequest.StopContext
	ugcLoadRequest.StopContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/")
	c.Abort()
	return
}

func deleteTenant(c *gin.Context) {

	session := sessions.Default(c)

	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)
	addMonitorAndDashboard(ugcLoadRequest)
	addTenants(ugcLoadRequest)

	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
		addMonitorAndDashboard(ugcLoadRequest)
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	log.WithFields(log.Fields{
		"TenantContext": ugcLoadRequest.TenantContext,
	}).Info("TenantContext")

	if ugcLoadRequest.TenantContext == "" {
		ugcLoadRequest.TenantMissing = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	if stringInSlice(ugcLoadRequest.TenantContext, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantDelete = strings.Join(nonValidNamespaces, ",")
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
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
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}
	ugcLoadRequest.TenantDeleted = ugcLoadRequest.TenantContext
	ugcLoadRequest.TenantContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/")
	c.Abort()
	return

}

func upload(c *gin.Context) {

	session := sessions.Default(c)

	kubctlOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)
	addTenants(ugcLoadRequest)
	addMonitorAndDashboard(ugcLoadRequest)

	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	/*

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
			return c.HTML(http.StatusOK, "index.tmpl", ugcLoadRequest)
		}

	*/

	if stringInSlice(ugcLoadRequest.Context, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantName = strings.Join(nonValidNamespaces, ",")
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	fop := ugl.FileUploadOperations{Context: c}
	testPath := fop.ProcessJmeter()
	if testPath == "" {
		_ = fop.ProcessData()
		kubctlOps.RegisterClient()
		ugcLoadRequest.MissingJmeter = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	fop.ProcessData()

	nsExist := kubctlOps.CheckNamespaces(ugcLoadRequest.Context)
	if nsExist == false {
		clusterops := cluster.Operations{}
		awsRegion, awsAcntNumber := clusterops.DescribeCluster("ugctestgrid")
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
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
		}
		created, errNs := kubctlOps.CreateNamespace(ugcLoadRequest.Context)
		if created == false {
			log.WithFields(log.Fields{
				"err": errNs,
			}).Error("Unable to create namespace")
			ugcLoadRequest.GenericCreateTestMsg = errNs
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
		}
		policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/ugcupload-eks-jmeter-policy", awsAcntNumber)
		crtd, e := kubctlOps.CreateServiceaccount(ugcLoadRequest.Context, policyArn)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create ServiceAccount")
			ugcLoadRequest.GenericCreateTestMsg = e
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
		}
		crtd, e = kubctlOps.CreateJmeterMasterDeployment(ugcLoadRequest.Context, aan, awsRegion)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Master Deployment")
			ugcLoadRequest.GenericCreateTestMsg = e
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
		}
		crtd, e = kubctlOps.CreateJmeterSlaveService(ugcLoadRequest.Context)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Slave Service")
			ugcLoadRequest.GenericCreateTestMsg = e
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
		}
		crtd, e = kubctlOps.CreateJmeterSlaveDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes), aan, awsRegion)
		if crtd == false {
			log.WithFields(log.Fields{
				"err": e,
			}).Error("Unable To Create Jmeter Slave Deployment")
			ugcLoadRequest.GenericCreateTestMsg = e
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/")
			c.Abort()
			return
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
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	ugcLoadRequest.Success = fmt.Sprintf("Test %s was succesfully created for tenant[%s]", testPath, ugcLoadRequest.Context)
	ugcLoadRequest.NumberOfNodes = 0
	ugcLoadRequest.Context = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/")
	c.Abort()
	return
}

func SetNoCacheHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("******************* RESPONSE HEADERS")
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Next()
	}
}

func main() {
	// Gin instance
	r := gin.Default()

	binding.Validator = new(defaultValidator)
	gob.Register(UgcLoadRequest{})
	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	r.Use(SetNoCacheHeader())
	//e.Validator = &CustomValidator{validator: validator.New()}

	kubctlOps.Init()
	kubctlOps.RegisterClient()

	r.LoadHTMLGlob(props.MustGet("web") + "/templates/*")
	r.GET("/", func(c *gin.Context) {

		session := sessions.Default(c)
		var ugcLoadRequest UgcLoadRequest
		if ulr := session.Get("ugcLoadRequest"); ulr != nil {
			ugcLoadRequest = ulr.(UgcLoadRequest)
		} else {
			ugcLoadRequest = UgcLoadRequest{}
		}
		addMonitorAndDashboard(&ugcLoadRequest)
		addTenants(&ugcLoadRequest)
		c.HTML(http.StatusOK, "index.tmpl", ugcLoadRequest)
		session.Clear()
		if err := session.Save(); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Unable to save the session")
		}

	})
	r.Static("/script", props.MustGet("web"))
	r.POST("/start-test", upload)
	r.POST("/stop-test", stopTest)
	r.POST("/delete-tenant", deleteTenant)
	r.GET("/tenantReport", s3Tenants)
	r.POST("/genReport", generateReport)
	r.GET("/tenants", allTenants)

	s := &http.Server{
		Addr:         ":1323",
		Handler:      r,
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	// Start server
	s.ListenAndServe()
}
