package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
	cache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

var systemCache = cache.New(cache.NoExpiration, 100*time.Minute)
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
	ip := cnt.KubeOps.LoadBalancerIP("control", "admin-controller")
	log.WithFields(log.Fields{
		"ur.MonitorURL":    fmt.Sprintf("http://%s:4040", ip),
		"ur.ReportURL":     fmt.Sprintf("http://%s:80", ip),
		"ur.DashboardURL":  fmt.Sprintf("http://%s:3000", cnt.KubeOps.LoadBalancerIP("ugcload-reporter", "jmeter-grafana")),
		"ur.ChronografURL": fmt.Sprintf("http://%s:8888", cnt.KubeOps.LoadBalancerIP("ugcload-reporter", "jmeter-chronograf")),
	}).Info("MonitoringDetails")
	ur.MonitorURL = fmt.Sprintf("http://%s:4040", ip)
	ur.ReportURL = fmt.Sprintf("http://%s:80", ip)
	ur.DashboardURL = fmt.Sprintf("http://%s:3000", cnt.KubeOps.LoadBalancerIP("ugcload-reporter", "jmeter-grafana"))
	ur.ChronografURL = fmt.Sprintf("http://%s:8888", cnt.KubeOps.LoadBalancerIP("ugcload-reporter", "jmeter-chronograf"))
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

//ForceStopTest used to stop the test
func (cnt *Controller) ForceStopTest(c *gin.Context) {

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

	var failedToStop = make(map[string]interface{})
	hsnames := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcLoadRequest.StopContext)
	if len(hsnames) > 0 {
		errStr, resp := jm.KillMaster(hsnames[0])
		if resp == false {
			failedToStop["master"] = errStr
		}
		slaveIps := cnt.KubeOps.GetPodIpsForSlaves(ugcLoadRequest.StopContext)
		for _, ip := range slaveIps {
			err, resp := jm.KillSlave(ip)
			if resp == false {
				failedToStop[fmt.Sprintf("SLAVE-%s", ip)] = err
			}
		}

		if len(failedToStop) > 1 {
			log.WithFields(log.Fields{
				"Context": ugcLoadRequest.StopContext,
				"err":     errStr,
			}).Info("Unable to stop the test")
			failed, _ := json.Marshal(failedToStop)
			ugcLoadRequest.TennantNotStopped =
				fmt.Sprintf("Following did not stop: %s: maybe try again or delete the tenant", failed)
			session.Set("ugcLoadRequest", ugcLoadRequest)
			session.Save()
			c.Redirect(http.StatusMovedPermanently, "/update")
			c.Abort()
			return
		}
	} else {
		log.WithFields(log.Fields{
			"Context":   ugcLoadRequest.StopContext,
			"hostnames": strings.Join(hsnames, ","),
		}).Info("No master nodes found")
		ugcLoadRequest.TennantNotStopped =
			fmt.Sprintf("Master was not found: %s", strings.Join(hsnames, ","))
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	e, d := redisUtils.RemoveFromWaitingTests(ugcLoadRequest.StopContext)
	err, del := redisUtils.RemoveTenant(ugcLoadRequest.StopContext)
	if del == false || d == false {
		ugcLoadRequest.TennantNotStopped = fmt.Sprintf("Unable to remove tenant from redis tenants list: %s or  maybe from waitinglist for force stop", err, e)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return

	}

	var stopChannel chan string
	if x, found := systemCache.Get(ugcLoadRequest.Context); found {
		stopChannel = x.(chan string)
		stopChannel <- "stop"
	}

	ugcLoadRequest.TenantStopped = ugcLoadRequest.StopContext
	ugcLoadRequest.StopContext = ""
	session.Set("ugcLoadRequest", ugcLoadRequest)
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/update")
	c.Abort()
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

	e, d := redisUtils.RemoveFromWaitingTests(ugcLoadRequest.StopContext)
	err, del := redisUtils.RemoveTenant(ugcLoadRequest.StopContext)
	if del == false || d == false {
		ugcLoadRequest.TennantNotStopped = fmt.Sprintf("Unable to remove tenant from redis tenants list: %s or maybe from waitinglist", err, e)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return

	}

	var stopChannel chan string
	if x, found := systemCache.Get(ugcLoadRequest.Context); found {
		stopChannel = x.(chan string)
		stopChannel <- "stop"
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
		slaves, err, fnd := cnt.KubeOps.GetallJmeterSlavesStatus(ugcLoadRequest.Context)
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
	var foundDeadSlave bool
	var deadSlave []string
	if x, found := systemCache.Get(ugcloaddRequest.Context); found {

		stopChannel = x.(chan string)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		for {
			select {
			case message := <-stopChannel:
				log.WithFields(log.Fields{
					"message": message,
					"tenant":  ugcloaddRequest.Context,
				}).Info("Received stop message: ending monitoring")
				return
			case <-ctx.Done():
				if !foundDeadSlave {
					for _, slave := range slaves {
						yes := jm.IsSlaveRunning(slave)
						if yes == false {
							foundDeadSlave = true
							deadSlave = append(deadSlave, slave)
						}
					}
					cancel()
					duration := 10 * time.Second
					if foundDeadSlave {
						duration = 30 * time.Second
					}
					ctx, cancel = context.WithTimeout(context.Background(), duration)
				} else {
					break
				}
			}
		}
	}

	if foundDeadSlave {
		log.WithFields(log.Fields{
			"node(s)": strings.Join(deadSlave, ","),
			"Tenant":  ugcloaddRequest.Context,
		}).Error("Node(s) has failed restarting test")

		//Stopping any tests that are running
		hsnames := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcloaddRequest.Context)
		_, _ = jm.StopTestOnMaster(hsnames[0])

		//Waiting a few seonds for all slaves to be notified
		time.Sleep(5 * time.Second)

		before := fmt.Sprintf("Node Current Budget: xmx=%s, xms=%s, maxmetaspacesize=%s, cpu=%s, ram=%s",
			ugcloaddRequest.Xms, ugcloaddRequest.Xmx, ugcloaddRequest.MaxMetaspaceSize, ugcloaddRequest.CPU,
			ugcloaddRequest.RAM)
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

		err := fmt.Sprintf("%s: Node New confiuration: xmx=%s, xms=%s, maxmetaspacesize=%s, cpu=%s, ram=%s",
			before, ugcloaddRequest.Xms, ugcloaddRequest.Xmx, ugcloaddRequest.MaxMetaspaceSize, ugcloaddRequest.CPU,
			ugcloaddRequest.RAM)
		redisTenant := types.RedisTenant{Tenant: ugcloaddRequest.Context}
		redisTenant.Started = fmt.Sprintf("Restart - because Node(s)=%s could not be reached", strings.Join(deadSlave, ","))
		redisTenant.Errors = err
		redisUtils.AddTenant(redisTenant)
		go cnt.startTest(ugcloaddRequest, jm)
		//Release semaphore
		return

	}

}

func (cnt *Controller) waitForJmeterToStartRunningOnSlaves(slaves []string, jm jmeter.Jmeter) (allRunning bool, failed []string) {

	var running []string
	var temp = slaves
	count := 0
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	for {
		select {
		case <-ctx.Done():

			var notRunning []string
			for _, slave := range temp {
				if jm.IsSlaveRunning(slave) {
					running = append(running, slave)
					fmt.Println(fmt.Sprintf("Running:%s", slave))
				} else {
					fmt.Println(fmt.Sprintf("Not Running:%s", slave))
					notRunning = append(notRunning, slave)
				}
			}

			if len(running) == len(slaves) {
				fmt.Println("ALL SLAVES ARE RUNNING")
				allRunning = true
				return
			}
			count = count + 1
			if count == 6 {
				allRunning = false
				failed = notRunning
				return
			}
			temp = notRunning
			notRunning = []string{}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
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
		err, created := kubeAdmin.CreateTenantInfrastructure(ugcLoadRequest)
		if created == false {
			redisTenant.Started = "failed"
			redisTenant.Errors = err
			redisUtils.AddTenant(redisTenant)
			return
		}

	} else {
		exist := cnt.KubeOps.DoesDeploymentExist(ugcLoadRequest.Context, "jmeter-slave")
		if exist {
			redisTenant.Started = fmt.Sprintf("Scaling nodes to %d", ugcLoadRequest.NumberOfNodes)
			redisUtils.AddTenant(redisTenant)
			scaledError, scaled := cnt.KubeOps.ScaleDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes))
			if scaled == false {
				redisTenant.Started = "problems scaling nodes"
				redisTenant.Errors = scaledError
				redisUtils.AddTenant(redisTenant)
				return
			}
		} else {
			//Something went wrong before recreating
			kubeAdmin := admin.Admin{KubeOps: &cnt.KubeOps, RedisUtil: redisUtils}
			err, created := kubeAdmin.CreateTenantInfrastructure(ugcLoadRequest)
			if created == false {
				redisTenant.Started = "failed"
				redisTenant.Errors = err
				redisUtils.AddTenant(redisTenant)
				return
			}
		}
	}

	slavesStarted := cnt.waitForSlavesToStartRunning(ugcLoadRequest)
	ips := cnt.KubeOps.GetPodIpsForSlaves(ugcLoadRequest.Context)
	if slavesStarted && len(ips) > 0 {
		listOfHost := strings.Join(ips, ",")
		redisTenant.Started = fmt.Sprintf("Starting slaves: %s", listOfHost)
		redisUtils.AddTenant(redisTenant)
		e, uploaded := jm.SetupSlaves(ugcLoadRequest, ips)
		if uploaded == false {
			redisTenant.Started = "failed"
			redisTenant.Errors = e
			redisUtils.AddTenant(redisTenant)
			return
		}
		masters := cnt.KubeOps.GetHostNamesOfJmeterMaster(ugcLoadRequest.Context)
		redisTenant.Started = fmt.Sprintf("Waiting for slaves to start on master %s ", strings.Join(masters, ","))
		redisTenant.Errors = fmt.Sprintf("%s", listOfHost)
		redisUtils.AddTenant(redisTenant)
		running, failedSlaves := cnt.waitForJmeterToStartRunningOnSlaves(ips, jm)
		if running == true {
			redisTenant.Started = fmt.Sprintf("start master %s", strings.Join(masters, ","))
			redisUtils.AddTenant(redisTenant)
			for _, master := range masters {
				e, started := jm.StartMasterTest(master, ugcLoadRequest, listOfHost)

				//NOTE: Something odd is happening
				if started == false {
					redisTenant.Started = "Running: But something odd happened"
					redisTenant.Errors = e
				} else {
					redisTenant.Started = "Running"
				}

				mess, created := cnt.KubeOps.CreatePodDisruptionBudget(ugcLoadRequest)
				if created == false {
					redisTenant.Started = "Running: Problems Creating PodDisruptionBudget"
					redisTenant.Errors = mess
				}
			}
			redisUtils.AddTenant(redisTenant)
			return
		}

		redisTenant.Errors = fmt.Sprintf("Hosts:%s", strings.Join(failedSlaves, ","))
		redisTenant.Started = "Jmeter slaves have not been started"
		return

		//ch := make(chan string, 1)
		//systemCache.Set(ugcLoadRequest.Context, ch, cache.NoExpiration)
		//go cnt.monitorMe(ugcLoadRequest, ips, jm)

	} else {
		redisUtils.AddTenant(types.RedisTenant{Tenant: ugcLoadRequest.Context, Started: "failed", Errors: fmt.Sprintf("No slaves found: For [%s]", ugcLoadRequest.Context)})
		return
	}

}

//FailingNodes get all failling nodes
func (cnt *Controller) FailingNodes(c *gin.Context) {
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

	redisTenant, e, found := redisUtils.GetTenant(ugcLoadRequest.Context)
	if found == false {
		ugcLoadRequest.GenericCreateTestMsg = e
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
	e, added := redisUtils.AddTenant(redisTenant)
	if added == false {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Unable to add test to list of pending test [%s]: %s", ugcLoadRequest.Context, e)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	e, added = redisUtils.AddToListOfStarted(ugcLoadRequest.Context)
	if added == false {
		ugcLoadRequest.GenericCreateTestMsg = fmt.Sprintf("Unable to add test details to list of started tests [%s]: %s", ugcLoadRequest.Context, e)
		session.Set("ugcLoadRequest", ugcLoadRequest)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/update")
		c.Abort()
		return
	}

	fileName, dataFile, _, fndData := getFileFromContext("data", c)
	if fndData == false {
		fileName = ""
	}
	/*
		if fndData == false {
			redisTenant.Started = "failed"
			redisTenant.Errors = errData
			redisUtils.AddTenant(redisTenant)
			return
		}
	*/

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

func (cnt *Controller) JmeterSlaves(c *gin.Context) {
	tenant := c.Query("tenant")
	items, _, _ := cnt.KubeOps.GetallJmeterSlavesStatus(tenant)
	c.PureJSON(http.StatusOK, items)
}

func (cnt *Controller) Testoutput(c *gin.Context) {

	ip := c.Query("ip")
	target := fmt.Sprintf("http://%s:1007/test-output", ip)
	jm := jmeter.Jmeter{}
	resp, err := jm.MakeRequest(target)
	if err == nil {

		cp := strings.TrimLeft(resp.Header["Content-Disposition"][0], "'")
		cp = strings.TrimRight(cp, "'")
		_, i := utf8.DecodeRuneInString(cp)
		cptrim := cp[i:]
		c.Writer.Header().Set("Content-type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", cptrim)
		_, e := io.Copy(c.Writer, resp.Body)
		if e != nil {
			log.WithFields(log.Fields{
				"err": e.Error(),
				"ip":  ip,
			}).Error("Problems getting jmeter log files")
		}
		return
	}
	log.WithFields(log.Fields{
		"err": err.Error(),
		"ip":  ip,
	}).Error("Problems making request to get jmeter log files")
	c.Abort()
	return
}
