package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	aws "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/aws"
	jmeter "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/jmeter"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"
	admin "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes/admin"
	redis "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/redis"
	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"
	validate "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/validate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/magiconair/properties"
	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

var systemCache = cache.New(cache.NoExpiration, 100*time.Minute)
var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)
var redisUtils redis.Redis

func init() {
	redisUtils = redis.Redis{}
	redisUtils.Setup()
}

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

	envRun := os.Getenv("ADMIN_CONTROLLER_PORT_1323_TCP")
	if envRun != "" {
		var running []types.Tenant
		for _, t := range cnt.tenantStatus() {
			log.WithFields(log.Fields{
				"tenant": t,
			}).Info("Tenant")
			if t.Running == true {
				running = append(running, t)
			}
		}

		ur.RunningTests = running
	}

	t, _ := cnt.KubeOps.GetallTenants()
	var filtered []types.Tenant
	tbd, _, _ := redisUtils.FetchWaitingToBeDeleted()
	validator := validate.Validator{}
	for _, ten := range t {
		if !validator.StringInSlice(ten.Namespace, tbd) {
			filtered = append(filtered, ten)
		}
	}
	ur.AllTenants = filtered
}

func (cnt *Controller) tenantStatus() (tenants []types.Tenant) {
	cnt.KubeOps.RegisterClient()
	t, e := cnt.KubeOps.GetallTenants()

	if e != "" {
		return
	}

	nt := []types.Tenant{}

	jm := jmeter.Jmeter{}
	for _, tenant := range t {
		err, r := jm.IsRunning(tenant.PodIP)
		log.WithFields(log.Fields{
			"r": r,
			"e": err,
		}).Info("response from check if running")
		tenant.Running = r
		nt = append(nt, tenant)
	}
	tenants = nt
	return
}

//AllTenants fetches all the tenants
func (cnt *Controller) AllTenants(c *gin.Context) {

	request := types.UgcLoadRequest{}
	cnt.AddTenants(&request)
	c.PureJSON(http.StatusOK, request)
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

	jm := jmeter.Jmeter{}
	hsnames := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcLoadRequest.StopContext)
	errStr, resp := jm.StopTestOnMaster(hsnames[0])
	if resp == false {
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

	error, del := redisUtils.RemoveTenant(ugcLoadRequest.StopContext)
	if del == false {
		ugcLoadRequest.TennantNotStopped = fmt.Sprintf("Unable to remove tenant from redis: %s", error)
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

func (cnt *Controller) removeTenant(tenant string) {
	deleted, errStr := cnt.KubeOps.DeleteServiceAccount(tenant)
	if deleted == false {
		t := types.RedisTenant{Tenant: tenant, Errors: errStr, Started: "problems deleting tenant"}
		redisUtils.AddToWaitingForDelete(t)
	} else {

		redisUtils.RemoveTenant(tenant)
		redisUtils.RemoveFromWaitingTests(tenant)
		redisUtils.RemoveTenantDelete(tenant)
		redisUtils.RemoveFromWaitingForDelete(tenant)
	}
}

//DeleteTenant used for deleting the tenant
func (cnt *Controller) DeleteTenant(c *gin.Context) {

	session := sessions.Default(c)
	cnt.KubeOps.RegisterClient()
	validator := validate.Validator{Context: c}

	ugcLoadRequest := new(types.UgcLoadRequest)
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

	go cnt.removeTenant(ugcLoadRequest.TenantContext)

	redisTenant := types.RedisTenant{Tenant: ugcLoadRequest.TenantContext, Started: "Started Deleting"}
	redisUtils.AddToWaitingForDelete(redisTenant)
	redisUtils.BeingDeleted(ugcLoadRequest.TenantContext)
	ugcLoadRequest.TenantDeleted = ugcLoadRequest.TenantContext
	ugcLoadRequest.TenantContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
	return

}

func getFileFromContext(name string, context *gin.Context) (fileName string, file *multipart.File, error string, opened bool) {

	fh, err := context.FormFile(name)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Unable to get the test data from the form")
		error = err.Error()
		opened = false
		return
	}
	fileName = fh.Filename

	log.WithFields(log.Fields{
		"size": fh.Size,
		"name": name,
	}).Error("Size of files")

	if fh != nil {
		f, err := fh.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err.Error(),
				"filename": fh.Filename,
			}).Error("Could not open the file")
			error = err.Error()
			opened = false
			return
		}
		opened = true
		file = &f
		return
	}

	error = "multipart header null"
	opened = false
	return
}

//waitForSlavesToStartRunning used to wait for all slaves to start running
func (cnt *Controller) waitForSlavesToStartRunning(ugcLoadRequest types.UgcLoadRequest) (running bool) {

	numNodes := ugcLoadRequest.NumberOfNodes + 1
	redisTenant := types.RedisTenant{}
	redisTenant.Tenant = ugcLoadRequest.Context
	redisTenant.Started = fmt.Sprintf("Waiting for maximum of 30 minutes for %d nodes to start", numNodes)
	redisUtils.AddTenant(redisTenant)

	count := 0
	found := 0
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {

		fmt.Println("tock")
		slaves, err, fnd := cnt.KubeOps.GetallJmeterSlaves(ugcLoadRequest.Context)
		if fnd == false {
			redisTenant.Errors = err
			redisTenant.Started = fmt.Sprintf("Only %d out of %d nodes were started", found, numNodes)
			redisUtils.AddTenant(redisTenant)
			return
		}

		for _, slave := range slaves {
			if strings.EqualFold(fmt.Sprintf("%s", slave.Phase), "Running") {
				found = found + 1
			}

		}

		redisTenant.Tenant = ugcLoadRequest.Context
		redisTenant.Started = fmt.Sprintf("%d out of %d nodes were started", found, numNodes)
		redisUtils.AddTenant(redisTenant)
		if found > numNodes {
			running = true
			return
		}

		if count == 360 {
			running = false
			return
		}
		count = count + 1
	}
	return
}

func (cnt *Controller) monitorMe(ugcloaddRequest types.UgcLoadRequest, slaves []string, jm jmeter.Jmeter) {

	var stopChannel chan string
	if x, found := systemCache.Get(ugcloaddRequest.Context); found {
		stopChannel = x.(chan string)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		for {
			select {
			case message := <-stopChannel:
				log.WithFields(log.Fields{
					"message": message,
					"tenant":  ugcloaddRequest.Context,
				}).Info("Received stop message: ending monitoring")
				return
			case <-ctx.Done():
				//Grab hold of the semaphore
				for _, slave := range slaves {
					yes := jm.IsSlaveRunning(slave)
					if yes == false {
						log.WithFields(log.Fields{
							"node":   slave,
							"tenant": ugcloaddRequest.Context,
						}).Error("Node has failed restarting test")
						//Stopping any tests that are running
						hsnames := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcloaddRequest.Context)
						_, _ = jm.StopTestOnMaster(hsnames[0])

						//Waiting a few seonds for all slaves to be notified
						time.Sleep(5 * time.Second)

						//Restart Test:
						cpu, _ := strconv.Atoi(ugcloaddRequest.CPU)
						ram, _ := strconv.Atoi(ugcloaddRequest.RAM)
						mmss, _ := strconv.Atoi(ugcloaddRequest.MaxMetaspaceSize)
						xms, _ := strconv.Atoi(ugcloaddRequest.Xms)
						xmx, _ := strconv.Atoi(ugcloaddRequest.Xmx)
						ugcloaddRequest.CPU = strconv.Itoa(cpu + 1)
						ugcloaddRequest.RAM = strconv.Itoa(ram + 1)
						ugcloaddRequest.MaxMetaspaceSize = strconv.Itoa(mmss + 256)

						ugcloaddRequest.Xms = strconv.Itoa(xms + 1)
						ugcloaddRequest.Xmx = strconv.Itoa(xmx + 1)
						cancel()
						go cnt.startTest(ugcloaddRequest, jm)
						//Release semaphore
						return
					}
				}
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
			}
		}
	}

}

func (cnt *Controller) startTest(ugcLoadRequest types.UgcLoadRequest, jm jmeter.Jmeter) {
	cnt.KubeOps.RegisterClient()
	redisTenant := types.RedisTenant{Tenant: ugcLoadRequest.Context}

	redisTenant.Started = "Checking to see if namespace exist"
	redisUtils.AddTenant(redisTenant)
	nsExist := cnt.KubeOps.CheckNamespaces(ugcLoadRequest.Context)
	if nsExist == false {
		redisTenant.Started = "Creating tenant infrastructure: This can take several minutes"
		redisUtils.AddTenant(redisTenant)
		//Creating tenant infrastructure
		kubeAdmin := admin.Admin{KubeOps: &cnt.KubeOps, RedisUtil: redisUtils}
		error, created := kubeAdmin.CreateTenantInfrastructure(ugcLoadRequest)
		if created == false {
			redisTenant.Started = "failed"
			redisTenant.Errors = error
			redisUtils.AddTenant(redisTenant)
			return
		}

	} else {
		redisTenant.Started = fmt.Sprintf("Scaling nodes to %d", ugcLoadRequest.NumberOfNodes)
		redisUtils.AddTenant(redisTenant)
		scaledError, scaled := cnt.KubeOps.ScaleDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes))
		if scaled == false {
			redisTenant.Started = "problems scaling nodes"
			redisTenant.Errors = scaledError
			redisUtils.AddTenant(redisTenant)
			return
		}
	}

	slavesStarted := cnt.waitForSlavesToStartRunning(ugcLoadRequest)
	ips := cnt.KubeOps.GetPodIpsForSlaves(ugcLoadRequest.Context)
	if slavesStarted && len(ips) > 0 {
		listOfHost := strings.Join(ips, ",")
		redisTenant.Started = fmt.Sprintf("Starting slaves: %s", listOfHost)
		redisUtils.AddTenant(redisTenant)
		error, uploaded := jm.SetupSlaves(ugcLoadRequest, ips)
		if uploaded == false {
			redisTenant.Started = "failed"
			redisTenant.Errors = error
			redisUtils.AddTenant(redisTenant)
			return
		}
		masters := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcLoadRequest.Context)
		redisTenant.Started = fmt.Sprintf("start master %s", strings.Join(masters, ","))
		redisUtils.AddTenant(redisTenant)
		time.Sleep(2 * time.Minute)
		for _, master := range masters {
			e, started := jm.StartMasterTest(master, ugcLoadRequest, listOfHost)
			tenant := types.RedisTenant{Tenant: ugcLoadRequest.Context}
			if started == false {
				tenant.Started = "Unable To Start"
				tenant.Errors = e
			} else {
				tenant.Started = "Running"
			}
			redisUtils.AddTenant(tenant)
		}

		ch := make(chan string, 1)
		systemCache.Set(ugcLoadRequest.Context, ch, cache.NoExpiration)
		go cnt.monitorMe(ugcLoadRequest, ips, jm)

	} else {
		redisUtils.AddTenant(types.RedisTenant{Tenant: ugcLoadRequest.Context, Started: "failed", Errors: fmt.Sprintf("No slaves found: Test for started for [%s]", ugcLoadRequest.Context)})
		return
	}

}

//FallingNodes get all failling nodes
func (cnt *Controller) FallingNodes(c *gin.Context) {
	cnt.KubeOps.RegisterClient()
	nodes, _ := cnt.KubeOps.GetAlFailingNodes()
	c.PureJSON(http.StatusOK, nodes)
}

//TestStatus used to get the status test
func (cnt *Controller) TestStatus(c *gin.Context) {

	tenants, _, _ := redisUtils.FetchWaitingTests()
	var redistTenants []types.RedisTenant
	for _, tenant := range tenants {
		rt, _, found := redisUtils.GetTenant(tenant)
		if found {
			if len(rt.Tenant) > 0 {
				redistTenants = append(redistTenants, rt)
			}
		}
	}

	tenants, _, _ = redisUtils.FetchWaitingToBeDeleted()
	var deletedTenants []types.RedisTenant
	for _, tenant := range tenants {
		rt, _, found := redisUtils.GetTenantFromDelete(tenant)
		if found {
			deletedTenants = append(deletedTenants, rt)
		}
	}

	ts := types.TestStatus{}
	ts.Started = redistTenants
	ts.BeingDeleted = deletedTenants
	c.PureJSON(http.StatusOK, ts)

}

//StartTest basicall does everything to start a test
func (cnt *Controller) StartTest(c *gin.Context) {

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

	log.WithFields(log.Fields{
		"ugcLoadRequest.Context":            ugcLoadRequest.Context,
		"ugcLoadRequest.NumberOfNodes":      ugcLoadRequest.NumberOfNodes,
		"ugcLoadRequest.BandWidthSelection": ugcLoadRequest.BandWidthSelection,
		"ugcLoadRequest.Jmeter":             ugcLoadRequest.Jmeter,
		"ugcLoadRequest.Data":               ugcLoadRequest.Data,
		"ugcLoadRequest.Xmx":                ugcLoadRequest.Xmx,
		"ugcLoadRequest.Xms":                ugcLoadRequest.Xms,
		"ugcLoadRequest.CPU":                ugcLoadRequest.CPU,
		"ugcLoadRequest.RAM":                ugcLoadRequest.RAM,
		"ugcLoadRequest.MaxMetaspaceSize":   ugcLoadRequest.MaxMetaspaceSize,
	}).Info("Request Properties")

	redisTenant, error, found := redisUtils.GetTenant(ugcLoadRequest.Context)
	if found == false {
		ugcLoadRequest.GenericCreateTestMsg = error
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	if !strings.EqualFold(redisTenant.Tenant, "") {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Waiting for [%s] to start: [%s] : [%s]", ugcLoadRequest.Context, redisTenant.Started, redisTenant.Errors)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return

	}

	beingDeleted, _, _ := redisUtils.CheckIfBeingDeleted(ugcLoadRequest.Context)
	if beingDeleted {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Need to wait for tenant [%s] to be deleted ", ugcLoadRequest.Context)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	redisTenant = types.RedisTenant{Tenant: ugcLoadRequest.Context, Started: "not-yet"}
	error, added := redisUtils.AddTenant(redisTenant)
	if added == false {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Unable to add test to list of pending test [%s]: %s", ugcLoadRequest.Context, error)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	error, added = redisUtils.AddToListOfStarted(ugcLoadRequest.Context)
	if added == false {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Unable to add test details to list of started tests [%s]: %s", ugcLoadRequest.Context, error)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	fileName, dataFile, errData, fndData := getFileFromContext("data", c)
	if fndData == false {
		redisTenant.Started = "failed"
		redisTenant.Errors = errData
		redisUtils.AddTenant(redisTenant)
		return
	}

	_, fileJmeter, errJmeter, fndJmeter := getFileFromContext("jmeter", c)
	if fndJmeter == false {
		redisTenant.Started = "failed"
		redisTenant.Errors = errJmeter
		redisUtils.AddTenant(redisTenant)
		return
	}

	jm := jmeter.Jmeter{Fop: ugl.FileUploadOperations{Context: c},
		Context: c, Data: dataFile, JmeterScript: fileJmeter, DataFilename: fileName}

	go cnt.startTest(*ugcLoadRequest, jm)

	responseContext := types.UgcLoadRequest{}
	responseContext.Success = fmt.Sprintf("Test for [%s] has been initiated", ugcLoadRequest.Context)
	session.Set("ugcLoadRequest", responseContext)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
	return
}
