package kubernetes

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	cluster "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/cluster"
	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/kubernetes"

	"strconv"

	"github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"

	"sync"
)

//Admin used to perform administrative operations on the cluster
type Admin struct {
	KubeOps *kubernetes.Operations
}

//DeployMaster used to deploy the master on kubernetes
func (admin *Admin) DeployMaster(ugcLoadRequest types.UgcLoadRequest,
	aan int64,
	awsRegion string,
	message sync.Map,
	wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	crtd, e := admin.KubeOps.CreateJmeterMasterDeployment(ugcLoadRequest.Context, aan, awsRegion)
	if crtd == false {
		log.WithFields(log.Fields{
			"err": e,
		}).Error("Unable To Create Jmeter Master Deployment")
		message.Store("masterDeploymentFailure", e)
	}
}

//DeploySlaveService usedf to the deploy the jmeter slaves on kubernetes
func (admin *Admin) DeploySlaveService(ugcLoadRequest types.UgcLoadRequest, message sync.Map, wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	crtd, e := admin.KubeOps.CreateJmeterSlaveService(ugcLoadRequest.Context)
	if crtd == false {
		log.WithFields(log.Fields{
			"err": e,
		}).Error("Unable To Create Jmeter Slave Service")
		message.Store("masterDeploymentFailure", e)
	}

}

//DeploySlavePods used to create the deployment for the slaves
func (admin *Admin) DeploySlavePods(ugcLoadRequest types.UgcLoadRequest, aan int64, awsRegion string, message sync.Map, wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	crtd, e := admin.KubeOps.CreateJmeterSlaveDeployment(ugcLoadRequest.Context, int32(ugcLoadRequest.NumberOfNodes), aan, awsRegion)
	if crtd == false {
		log.WithFields(log.Fields{
			"err": e,
		}).Error("Unable To Create Jmeter Slave Deployment")
		message.Store("masterDeploymentFailure", e)
	}
}

//CreateTenantInfrastructure used to create the infrastructure for the tenant
func (admin *Admin) CreateTenantInfrastructure(ugcLoadRequest types.UgcLoadRequest) (error string, result bool) {
	admin.KubeOps.RegisterClient()
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
		error = err.Error()
		result = false
		return
	}

	created, errNs := admin.KubeOps.CreateNamespace(ugcLoadRequest.Context)
	if created == false {
		log.WithFields(log.Fields{
			"err": errNs,
		}).Error("Unable to create namespace")
		error = errNs
		result = false
		return
	}
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/ugcupload-eks-jmeter-policy", awsAcntNumber)
	crtd, e := admin.KubeOps.CreateServiceaccount(ugcLoadRequest.Context, policyArn)
	if crtd == false {
		log.WithFields(log.Fields{
			"err": e,
		}).Error("Unable To Create ServiceAccount")
		error = e
		result = false
		return
	}

	var wg sync.WaitGroup
	message := sync.Map{}

	go admin.DeployMaster(ugcLoadRequest, aan, awsRegion, message, wg)
	go admin.DeploySlaveService(ugcLoadRequest, message, wg)
	go admin.DeploySlavePods(ugcLoadRequest, aan, awsRegion, message, wg)
	wg.Wait()

	responses := make(map[interface{}]interface{})
	message.Range(func(k, v interface{}) bool {
		responses[k] = v
		return true
	})

	if len(responses) != 0 {
		b := new(bytes.Buffer)
		for key, value := range responses {
			fmt.Fprintf(b, "%s:\"%s\"\n", key, value)
		}
		dr, err := admin.KubeOps.DeleteServiceAccount(ugcLoadRequest.Context)
		if dr == false {
			fmt.Fprintf(b, "%s:\"%s\"\n", "UnableToDeleteServiceAccountAfterDeploymentFailure", err)
		}
		error = b.String()
		result = false
	}

	result = true
	return
}
