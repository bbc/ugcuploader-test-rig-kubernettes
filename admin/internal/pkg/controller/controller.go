package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	aws "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/aws"
	cluster "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/cluster"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin/binding"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)
var nonValidNamespaces = []string{"control", "default", "kube-node-lease", "kube-public", "kube-system", "ugcload-reporter", "weave"}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//Controller the web control layer
type Controller struct {
	KubeOps kubernetes.Operations
	S3      aws.S3Operations
}

//UgcLoadRequest This is used to map to the form data.. seems to only work with firefox
type UgcLoadRequest struct {
	Context              string `json:"context" form:"context" validate:"required"`
	NumberOfNodes        int    `json:"numberOfNodes" form:"numberOfNodes" validate:"numeric,min=1"`
	BandWidthSelection   string `json:"bandWidthSelection" form:"bandWidthSelection" validate:"required"`
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

//AddMonitorAndDashboard adds the url of the dashboard and monitor the request
func (cnt *Controller) AddMonitorAndDashboard(ur *UgcLoadRequest) {
	cnt.KubeOps.RegisterClient()
	ip := cnt.KubeOps.LoadBalancerIP("control")
	ur.MonitorURL = fmt.Sprintf("http://%s:4040", ip)
	ur.ReportURL = fmt.Sprintf("http://%s:80", ip)
	ur.DashboardURL = fmt.Sprintf("http://%s:3000", cnt.KubeOps.LoadBalancerIP("ugcload-reporter"))
}

//AddTenants adds a list of tenants to the request
func (cnt *Controller) AddTenants(ur *UgcLoadRequest) {
	cnt.KubeOps.RegisterClient()
	ur.TenantList, _ = cnt.S3.GetBucketItems("ugcupload-jmeter", "", 0)
}

//AllTenants fetches all the tenants
func (cnt *Controller) AllTenants(c *gin.Context) {

	cnt.KubeOps.RegisterClient()
	t, e := cnt.KubeOps.GetallTenants()

	if e != "" {
		c.String(http.StatusBadRequest, "Unable to fetch tenants")
	}

	nt := []kubernetes.Tenant{}

	for _, tenant := range t {
		r, e := cnt.KubeOps.CheckIfRunningJava(tenant.Namespace, tenant.Name)
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

func (cnt *Controller) GenerateReport(c *gin.Context) {

	tenant, _ := c.GetPostForm("tenant")
	data, _ := c.GetPostForm("data")

	var items []string
	for _, d := range strings.Split(data, ",") {
		items = append(items, fmt.Sprintf("%s=%s", tenant, d))
	}
	cnt.KubeOps.RegisterClient()
	_, e := cnt.KubeOps.GenerateReport(strings.Join(items[:], ","))
	c.String(http.StatusOK, e)
}

func (cnt *Controller) S3Tenants(c *gin.Context) {

	type Items struct {
		Date string `json:"date"`
	}
	tenant, _ := c.GetQuery("tenant")

	var my []Items
	items, _ := cnt.S3.GetBucketItems("ugcupload-jmeter", fmt.Sprintf("%s/", tenant), 1)
	for _, item := range items {
		it := Items{Date: item}
		my = append(my, it)
	}
	c.JSON(http.StatusOK, &my)
	return
}

//StopTest used to stop the test
func (cnt *Controller) StopTest(c *gin.Context) {

	session := sessions.Default(c)
	cnt.KubeOps.RegisterClient()

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

	deleted, errStr := cnt.KubeOps.StopTest(ugcLoadRequest.StopContext)
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

func (cnt *Controller) DeleteTenant(c *gin.Context) {

	session := sessions.Default(c)

	cnt.KubeOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)

	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
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

	deleted, errStr := cnt.KubeOps.DeleteServiceAccount(ugcLoadRequest.TenantContext)
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

//Upload basicall does everything to start a test
func (cnt *Controller) Upload(c *gin.Context) {

	session := sessions.Default(c)

	cnt.KubeOps.RegisterClient()
	ugcLoadRequest := new(UgcLoadRequest)

	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	if len(strings.TrimSpace(ugcLoadRequest.Context)) < 3 {
		ugcLoadRequest.MissingTenant = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return

	}

	if ugcLoadRequest.NumberOfNodes < 1 {
		ugcLoadRequest.MissingNumberOfNodes = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return

	}

	if stringInSlice(ugcLoadRequest.Context, nonValidNamespaces) {
		ugcLoadRequest.InvalidTenantName = strings.Join(nonValidNamespaces, ",")
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	nsExist := cnt.KubeOps.CheckNamespaces(ugcLoadRequest.Context)
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
		created, errNs := cnt.KubeOps.CreateNamespace(ugcLoadRequest.Context)
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
		crtd, e := cnt.KubeOps.CreateServiceaccount(ugcLoadRequest.Context, policyArn)
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
		crtd, e = cnt.KubeOps.CreateJmeterMasterDeployment(ugcLoadRequest.Context, aan, awsRegion)
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
		crtd, e = cnt.KubeOps.CreateJmeterSlaveService(ugcLoadRequest.Context)
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
		crtd, e = cnt.KubeOps.CreateJmeterSlaveDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes), aan, awsRegion)
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
	fop := ugl.FileUploadOperations{Context: c}
	testPath := fop.ProcessJmeter()
	if testPath == "" {
		cnt.KubeOps.RegisterClient()
		ugcLoadRequest.MissingJmeter = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/")
		c.Abort()
		return
	}

	cnt.KubeOps.ScaleDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes))
	cnt.KubeOps.WaitForPodsToStart(ugcLoadRequest.Context, ugcLoadRequest.NumberOfNodes+1)
	hostnames := cnt.KubeOps.GetHostEndpoints(ugcLoadRequest.Context)
	fmt.Println(fmt.Sprint("host names=%s", hostnames))

	for _, hn := range hostnames {
		dataURI := fmt.Sprintf("http://%s:1007/data", hn)
		fop.ProcessData(dataURI)
		jmeterURI := fmt.Sprintf("http://%s:1007/jmeter-props", hn)
		fop.UploadJmeterProps(jmeterURI, ugcLoadRequest.BandWidthSelection)
		resp, err := http.Get(fmt.Sprintf("http://%s:1007/start-server", hn))
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err.Error(),
				"host": hn,
			}).Error("Problems starting the jmeter slave")
			// handle error
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		log.WithFields(log.Fields{
			"response": string(body),
		}).Info("Response from starting the jmeter slave")
	}

	s, er := cnt.KubeOps.StartTest(testPath, ugcLoadRequest.Context, ugcLoadRequest.BandWidthSelection, ugcLoadRequest.NumberOfNodes)

	if s == false {
		cnt.KubeOps.RegisterClient()
		ugcLoadRequest.GenericCreateTestMsg = er
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
