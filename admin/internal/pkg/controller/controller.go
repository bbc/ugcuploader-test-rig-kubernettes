package controller

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	aws "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/aws"
	cluster "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/cluster"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"
	validate "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/validate"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin/binding"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//Controller the web control layer
type Controller struct {
	KubeOps kubernetes.Operations
	S3      aws.S3Operations
}

//AddMonitorAndDashboard adds the url of the dashboard and monitor the request
func (cnt *Controller) AddMonitorAndDashboard(ur *types.UgcLoadRequest) {
	cnt.KubeOps.RegisterClient()
	ip := cnt.KubeOps.LoadBalancerIP("control")
	ur.MonitorURL = fmt.Sprintf("http://%s:4040", ip)
	ur.ReportURL = fmt.Sprintf("http://%s:80", ip)
	ur.DashboardURL = fmt.Sprintf("http://%s:3000", cnt.KubeOps.LoadBalancerIP("ugcload-reporter"))
}

//AddTenants adds a list of tenants to the request
func (cnt *Controller) AddTenants(ur *types.UgcLoadRequest) {
	cnt.KubeOps.RegisterClient()
	ur.TenantList, _ = cnt.S3.GetBucketItems("ugcupload-jmeter", "", 0)

	var running []types.Tenant
	for _, t := range cnt.tenantStatus() {
		fmt.Println(fmt.Sprintf("%w runnn-----", t))
		if t.Running == true {
			running = append(running, t)
		}
	}

	ur.RunningTests = running

	t, _ := cnt.KubeOps.GetallTenants()
	ur.AllTenants = t
}

func (cnt *Controller) tenantStatus() (tenants []types.Tenant) {
	cnt.KubeOps.RegisterClient()
	t, e := cnt.KubeOps.GetallTenants()

	if e != "" {
		return
	}

	nt := []types.Tenant{}

	for _, tenant := range t {
		r, e := cnt.KubeOps.CheckIfRunningJava(tenant.Namespace, tenant.Name)
		if len(e) < 1 && len(r) < 1 {
			tenant.Running = true
		} else if len(e) > 0 || len(r) < 1 {
			tenant.Running = false
		} else {
			tenant.Running = true
		}
		nt = append(nt, tenant)
	}

	tenants = nt
	return
}

//AllTenants fetches all the tenants
func (cnt *Controller) AllTenants(c *gin.Context) {

	tenants := cnt.tenantStatus()
	c.PureJSON(http.StatusOK, tenants)
}

//GenerateReport used for generating the jmeter reports
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

//S3Tenants used to get all the tenants in the s3 bucket
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

	ugcLoadRequest := new(types.UgcLoadRequest)

	if err := c.ShouldBindWith(&ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	log.WithFields(log.Fields{
		"StopContext": ugcLoadRequest.StopContext,
	}).Info("StopContext")

	validator := validate.Validator{Context: c}

	if validator.ValidateStopTest(ugcLoadRequest) == false {
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
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
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	ugcLoadRequest.TenantStopped = ugcLoadRequest.StopContext
	ugcLoadRequest.StopContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
	return
}

//DeleteTenant used for deleting the tenant
func (cnt *Controller) DeleteTenant(c *gin.Context) {

	session := sessions.Default(c)

	cnt.KubeOps.RegisterClient()
	ugcLoadRequest := new(types.UgcLoadRequest)
	validator := validate.Validator{Context: c}
	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	log.WithFields(log.Fields{
		"TenantContext": ugcLoadRequest.TenantContext,
	}).Info("TenantContext")

	ugcLoadRequest.TenantContext = strings.TrimSpace(ugcLoadRequest.TenantContext)
	if validator.ValidateTenantDelete(ugcLoadRequest) == false {
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
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
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}
	ugcLoadRequest.TenantDeleted = ugcLoadRequest.TenantContext
	ugcLoadRequest.TenantContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
	return

}

//Upload basicall does everything to start a test
func (cnt *Controller) Upload(c *gin.Context) {

	session := sessions.Default(c)

	validator := validate.Validator{Context: c}

	cnt.KubeOps.RegisterClient()
	ugcLoadRequest := new(types.UgcLoadRequest)
	if err := c.ShouldBindWith(ugcLoadRequest, binding.Form); err != nil {
		ugcLoadRequest.ProblemsBinding = true
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	ugcLoadRequest.Context = strings.TrimSpace(ugcLoadRequest.Context)
	if validator.ValidateUpload(ugcLoadRequest) == false {
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
			c.Redirect(http.StatusMovedPermanently, "/update")
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
		c.Redirect(http.StatusMovedPermanently, "/update")
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
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:1007/start-server", hn), nil)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err.Error(),
				"host": hn,
			}).Error("Problems creating the request")
			// handle error
		} else {

			client := &http.Client{}
			resp, errResp := client.Do(req)
			if errResp != nil {
				log.WithFields(log.Fields{
					"err":  errResp.Error(),
					"host": hn,
				}).Error("Problems creating the request")

			} else {

				var bodyContent []byte
				resp.Body.Read(bodyContent)
				resp.Body.Close()
				log.WithFields(log.Fields{
					"response": string(bodyContent),
				}).Info("Response from starting the jmeter slave")
			}
		}

	}
	listOfHost := strings.Join(hostnames, ",")

	if len(listOfHost) > 0 {
		s, er := cnt.KubeOps.StartTest(testPath, ugcLoadRequest.Context, listOfHost)
		if s == false {
			cnt.KubeOps.RegisterClient()
			ugcLoadRequest.GenericCreateTestMsg = er
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/update")
			c.Abort()
			return
		}
	} else {
		cnt.KubeOps.RegisterClient()
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("No slaves found: Test for started for [%s]", ugcLoadRequest.Context)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	ugcLoadRequest.Success = fmt.Sprintf("Test %s was succesfully created for tenant[%s]", testPath, ugcLoadRequest.Context)
	ugcLoadRequest.NumberOfNodes = 0
	ugcLoadRequest.Context = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
	return
}
