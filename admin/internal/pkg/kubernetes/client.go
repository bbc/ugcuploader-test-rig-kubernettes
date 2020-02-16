package kubernetes

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	shellExec "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/exec"
	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//autoscaling "k8s.io/api/autoscaling/v1"
)

//Operations used for communicating with kubernetics api
type Operations struct {
	ClientSet *kubernetes.Clientset
	Config    *rest.Config
	TestPath  string
	Tenant    string
	Bandwidth string
	Nodes     string
}

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//Init init
func (kop *Operations) Init() (success bool) {

	if os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Errorf("Problems getting credentials")
			success = false
		} else {
			kop.Config = config
			success = true
		}

	} else {
		if kop.Config == nil {
			var kubeconfig *string
			if home := homeDir(); home != "" {
				kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			} else {
				kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			}
			flag.Parse()

			// use the current context in kubeconfig
			config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Errorf("Unable to initialize kubeconfig")
				success = false
			} else {
				kop.Config = config
				success = true
			}
		}
	}
	return
}

func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }

//DeleteDeployment use to delete the deployment
func (kop *Operations) DeleteDeployment(ns string, deployment string) (deleted bool) {

	dpf := metav1.DeletePropagationForeground
	options := &metav1.DeleteOptions{
		GracePeriodSeconds: int64Ptr(int64(0)),
		PropagationPolicy:  &dpf,
	}

	err := kop.ClientSet.AppsV1().Deployments(ns).Delete(deployment, options)
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err.Error(),
			"namespace": ns,
		}).Errorf("Problems deleting")
		deleted = false
		return
	}
	deleted = true
	return
}

//DoesDeploymentExist checks to see if the deployment exists
func (kop *Operations) DoesDeploymentExist(ns string, deployment string) (exist bool) {

	replicaSet, err := kop.ClientSet.AppsV1().Deployments(ns).Get(deployment, metav1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err.Error(),
			"namespace": ns,
		}).Errorf("Problem getting the scale")
		exist = false
		return
	}
	if replicaSet == nil {
		exist = false
		return
	}
	exist = true
	return
}

//ScaleDeployment used to scale the jmeter slave
func (kop *Operations) ScaleDeployment(ns string, replica int32) (error string, scaled bool) {

	scale, err := kop.ClientSet.AppsV1().Deployments(ns).GetScale("jmeter-slave", metav1.GetOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err.Error(),
			"replica":   replica,
			"namespace": ns,
		}).Errorf("Problem getting the scale")
		error = err.Error()
		scaled = false
		return
	}

	if scale.Spec.Replicas != replica {
		scale.Spec.Replicas = replica
		deploymentsClient := kop.ClientSet.AppsV1().Deployments(ns)
		_, e := deploymentsClient.UpdateScale("jmeter-slave", scale)
		if e != nil {
			log.WithFields(log.Fields{
				"err":       e.Error(),
				"replica":   replica,
				"namespace": ns,
			}).Errorf("Problem updating number of replicas")
			error = e.Error()
			scaled = false
			return
		}
	}
	scaled = true
	return
}

//DeleteNamespace delete namespace
func (kop *Operations) DeleteNamespace(ns string) (deleted bool, err string) {
	deletePolicy := metav1.DeletePropagationForeground
	log.WithFields(log.Fields{
		"nameapce": ns,
	}).Info("Namespace to delete : %s", ns)
	if e := kop.ClientSet.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Errorf("Problem deleting namespace: %s", ns)
		deleted = false
		err = fmt.Sprintf("%s", e.Error())
	} else {
		deleted = true
	}
	return
}

//CreateNamespace create namespace
func (kop *Operations) CreateNamespace(ns string) (created bool, err string) {

	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	_, e := kop.ClientSet.CoreV1().Namespaces().Create(nsSpec)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Errorf("Problem creating namespace: %s", ns)
		created = false
		err = fmt.Sprintf("%w", e.Error())
	} else {
		created = true
	}
	return
}

//GetAlFailingNodes returns a list of nodes that have failed
func (kop *Operations) GetAlFailingNodes() (nodes []types.NodePhase, found bool) {
	actual := metav1.ListOptions{}
	var nodePhases []types.NodePhase
	res, e := kop.ClientSet.CoreV1().Nodes().List(actual)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems getting all nodes")
		found = false
		return
	}

	for _, item := range res.Items {

		if len(item.Spec.Taints) > 0 {

			first := true
			out := ""
			for _, taint := range item.Spec.Taints {

				if !first {
					out = "," + out
				} else {
					first = false
				}
				out = out + taint.Key + ":" + taint.Value + "|"
			}
			nodePhase := types.NodePhase{}
			var nodeConditions []types.NodeCondition
			nodePhase.Phase = out
			nodePhase.InstanceID = item.Labels["alpha.eksctl.io/instance-id"]
			nodePhase.Name = item.Name
			for _, condition := range item.Status.Conditions {
				con := types.NodeCondition{}
				con.Type = string(condition.Type)
				con.Status = string(condition.Status)
				con.LastHeartbeatTime = condition.LastHeartbeatTime.String()
				con.Reason = condition.Reason
				con.Message = condition.Message
				nodeConditions = append(nodeConditions, con)
			}
			nodePhase.NodeConditions = nodeConditions
			nodePhases = append(nodePhases, nodePhase)
			found = true
		}
	}
	nodes = nodePhases
	found = false
	return
}

//GetallJmeterSlavesStatus gets all the jmeter slaves
func (kop *Operations) GetallJmeterSlavesStatus(tenant string) (slvs []types.SlaveStatus, err string, found bool) {
	slaves := []types.SlaveStatus{}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"jmeter_mode": "slave"}}
	actual := metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}
	res, e := kop.ClientSet.CoreV1().Pods(tenant).List(actual)
	if e != nil {
		log.WithFields(log.Fields{
			"err":    e.Error(),
			"Tenant": tenant,
		}).Error("Problems getting all slaves")
		err = e.Error()
		found = false
		return
	} else {
		for _, item := range res.Items {
			slaves = append(slaves, types.SlaveStatus{Name: item.Name, Phase: string(item.Status.Phase), PodIP: string(item.Status.PodIP)})
		}

		if len(slaves) < 1 {
			log.WithFields(log.Fields{
				"err":    "maybe the selector is wrong",
				"Tenant": tenant,
			}).Error("Problems getting all slaves")
			err = "something abnormal happened"
			found = false
		}
		slvs = slaves
		found = true
	}
	return
}

//GetallTenants Retuns a list of tenants
func (kop *Operations) GetallTenants() (ts []types.Tenant, err string) {
	tenants := []types.Tenant{}
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"jmeter_mode": "master"}}
	actual := metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}
	res, e := kop.ClientSet.CoreV1().Pods("").List(actual)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems getting all namespaces")
		err = e.Error()
	} else {
		for _, item := range res.Items {
			tenants = append(tenants, types.Tenant{Name: item.Name, Namespace: item.Namespace, PodIP: item.Status.PodIP})
		}
		ts = tenants
	}
	return
}

//CreateTelegrafConfigMap the config map used by the telgraf sidecar
func (kop *Operations) CreateTelegrafConfigMap(ns string) (created bool, err string) {

	connfigmapsclient := kop.ClientSet.CoreV1().ConfigMaps(ns)

	hn := fmt.Sprintf("%s", ns)
	conf := `
	[global_tags]
		tenant = "%s"
  [agent]
	hostname = "%s-$HOSTNAME"
	  [[outputs.influxdb]]
		urls = ["http://influxdb-jmeter.ugcload-reporter.svc.cluster.local:8086"]
		skip_database_creation = false
		database = "jmeter-slaves"
		write_consistency = "any"
		timeout = "5s"

	[[inputs.jolokia2_agent]]
		urls = ["http://localhost:8778/jolokia"]
	
	[[inputs.jolokia2_agent.metric]]
		name  = "java_runtime"
		mbean = "java.lang:type=Runtime"
		paths = ["Uptime"]
	
	[[inputs.jolokia2_agent.metric]]
		name  = "java_memory"
		mbean = "java.lang:type=Memory"
		paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
	
	[[inputs.jolokia2_agent.metric]]
		name     = "java_garbage_collector"
		mbean    = "java.lang:name=*,type=GarbageCollector"
		paths    = ["CollectionTime", "CollectionCount"]
		tag_keys = ["name"]
	
	[[inputs.jolokia2_agent.metric]]
		name  = "java_last_garbage_collection"
		mbean = "java.lang:name=*,type=GarbageCollector"
		paths = ["LastGcInfo"]
		tag_keys = ["name"]
	
	[[inputs.jolokia2_agent.metric]]
		name  = "java_threading"
		mbean = "java.lang:type=Threading"
		paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]
	
	[[inputs.jolokia2_agent.metric]]
		name  = "java_class_loading"
		mbean = "java.lang:type=ClassLoading"
		paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]
	
	[[inputs.jolokia2_agent.metric]]
		name     = "java_memory_pool"
		mbean    = "java.lang:name=*,type=MemoryPool"
		paths    = ["Usage", "PeakUsage", "CollectionUsage"]
		tag_keys = ["name"]
	
	[[inputs.cgroup]]
	paths = [
	"/cgroup/memory",           # root cgroup
		"/cgroup/memory/child1",    # container cgroup
		"/cgroup/memory/child2/*",  # all children cgroups under child2, but not child2 itself
		]
	files = ["memory.*usage*", "memory.limit_in_bytes"]
	
	[[inputs.cgroup]]
	paths = [
	"/cgroup/cpu",              # root cgroup
	"/cgroup/cpu/*",            # all container cgroups
	"/cgroup/cpu/*/*",          # all children cgroups under each container cgroup
	]
	files = ["cpuacct.usage", "cpu.cfs_period_us", "cpu.cfs_quota_us"]
	
	[[inputs.filecount]]
		directories = ["/test-output"]
		name = "*"

	[[inputs.filecount]]
		directory = "/test-output"
		name = "*.plain"
	
	[[inputs.mem]]

	# Read metrics about cpu usage
	[[inputs.cpu]]
	## Whether to report per-cpu stats or not
	percpu = true
	## Whether to report total system cpu stats or not
	totalcpu = true
	## Comment this line if you want the raw CPU time metrics
	fielddrop = ["time_*"]
	
	
	# Read metrics about disk usage by mount point
	[[inputs.disk]]
	## By default, telegraf gather stats for all mountpoints.
	## Setting mountpoints will restrict the stats to the specified mountpoints.
	# mount_points = ["/"]
	
	## Ignore some mountpoints by filesystem type. For example (dev)tmpfs (usually
	## present on /run, /var/run, /dev/shm or /dev).
	ignore_fs = ["tmpfs", "devtmpfs"]
	
	
	# Read metrics about disk IO by device
	[[inputs.diskio]]
	## By default, telegraf will gather stats for all devices including
	## disk partitions.
	## Setting devices will restrict the stats to the specified devices.
	# devices = ["sda", "sdb"]
	## Uncomment the following line if you need disk serial numbers.
	# skip_serial_number = false
		
	# Get kernel statistics from /proc/stat
	[[inputs.kernel]]
	# no configuration
	
	
	# Read metrics about memory usage
	[[inputs.mem]]
	# no configuration
	
	
	# Get the number of processes and group them by status
	[[inputs.processes]]
	# no configuration
	
	
	# Read metrics about swap memory usage
	[[inputs.swap]]
	# no configuration
	
	
	# Read metrics about system load & uptime
	[[inputs.system]]
	# no configuration
	
	# Read metrics about network interface usage
	[[inputs.net]]
	# collect data only about specific interfaces
	# interfaces = ["eth0"]
	
	
	[[inputs.netstat]]
	# no configuration
	
	[[inputs.interrupts]]
	# no configuration
	
	[[inputs.linux_sysctl_fs]]
	# no configuration

  `
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "telegraf-config-map",
			Namespace: ns},
		Data: map[string]string{
			"telegraf.conf": fmt.Sprintf(conf, ns, hn),
		},
	}

	result, e := connfigmapsclient.Create(configmap)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems creating config map")
		created = false
		err = e.Error()
	} else {
		log.WithFields(log.Fields{
			"name": result.GetObjectMeta().GetName(),
		}).Info("Deployment succesful created config map")
		created = true
	}
	return
}

//CreatePodDisruptionBudget create a budget to prevent the pods from deleted
func (kop *Operations) CreatePodDisruptionBudget(ugcuploadRequest types.UgcLoadRequest) (mess string, created bool) {

	pdb := &v1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-disruption-budget", ugcuploadRequest.Context),
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"namespace": ugcuploadRequest.Context,
				},
			},
			MaxUnavailable: &intstr.IntOrString{IntVal: 0},
		},
	}

	_, e := kop.ClientSet.PolicyV1beta1().PodDisruptionBudgets(ugcuploadRequest.Context).Create(pdb)

	if e != nil {
		log.WithFields(log.Fields{
			"tenant": ugcuploadRequest.Context,
			"Error":  e.Error(),
		}).Error("Problems creating pod disruption budget(s)")
		mess = e.Error()
		created = false
		return
	}
	created = true
	return
}

//CreateJmeterSlaveDeployment creates deployment for jmeter slaves
func (kop *Operations) CreateJmeterSlaveDeployment(ugcuploadRequest types.UgcLoadRequest, nbrnodes int32, awsAcntNbr int64, awsRegion string) (created bool, err string) {

	values := []string{"slaves"}
	nodeSelectorRequirement := corev1.NodeSelectorRequirement{Key: "jmeter_mode", Operator: corev1.NodeSelectorOpIn, Values: values}
	nodeSelectorRequirements := []corev1.NodeSelectorRequirement{nodeSelectorRequirement}
	nodeSelectorTerm := corev1.NodeSelectorTerm{MatchExpressions: nodeSelectorRequirements}
	nodeSelectorTerms := []corev1.NodeSelectorTerm{nodeSelectorTerm}
	nodeSelector := &corev1.NodeSelector{NodeSelectorTerms: nodeSelectorTerms}
	nodeAffinity := &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: nodeSelector}
	affinity := &corev1.Affinity{NodeAffinity: nodeAffinity}

	configmapVolumeSource := &corev1.ConfigMapVolumeSource{
		LocalObjectReference: corev1.LocalObjectReference{Name: "telegraf-config-map"},
		Items: []corev1.KeyToPath{
			{
				Key:  "telegraf.conf",
				Path: "telegraf.conf",
			},
		},
	}

	emptyDirVolumeSource := &corev1.EmptyDirVolumeSource{
		Medium: corev1.StorageMediumDefault,
	}
	volumeSource := corev1.VolumeSource{
		ConfigMap: configmapVolumeSource,
	}

	testOuputVolumeSource := corev1.VolumeSource{
		EmptyDir: emptyDirVolumeSource,
	}

	cpuformat := fmt.Sprintf("%v", resource.NewMilliQuantity(500, resource.DecimalSI))
	memformat := fmt.Sprintf("%v", resource.NewQuantity(30*1024*1024, resource.BinarySI))
	resourcerequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuformat),
			corev1.ResourceMemory: resource.MustParse(memformat),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuformat),
			corev1.ResourceMemory: resource.MustParse(memformat),
		},
	}

	ram, _ := strconv.Atoi(ugcuploadRequest.RAM)
	cpu, _ := strconv.Atoi(ugcuploadRequest.CPU)

	cpuformatSlave := fmt.Sprintf("%v", resource.NewMilliQuantity(int64(cpu)*1000, resource.DecimalSI))
	memformatSlave := fmt.Sprintf("%v", resource.NewQuantity(int64(ram)*1024*1024*1024, resource.BinarySI))
	fmt.Println(fmt.Sprintf("----------------- REQUESTED MEMORY:%s", memformatSlave))
	resourcerequirementSlave := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuformatSlave),
			corev1.ResourceMemory: resource.MustParse(memformatSlave),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuformatSlave),
			corev1.ResourceMemory: resource.MustParse(memformatSlave),
		},
	}

	toleration := corev1.Toleration{Key: "jmeter_slave", Operator: corev1.TolerationOpExists, Value: "", Effect: corev1.TaintEffectNoSchedule}
	tolerations := []corev1.Toleration{toleration}
	deploymentsClient := kop.ClientSet.AppsV1().Deployments(ugcuploadRequest.Context)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "jmeter-slave",
			Labels: map[string]string{
				"jmeter_mode": "slave",
				"namespace":   ugcuploadRequest.Context,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(nbrnodes),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"jmeter_mode": "slave",
					"namespace":   ugcuploadRequest.Context,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jmeter_mode": "slave",
						"namespace":   ugcuploadRequest.Context,
					},
				},
				Spec: corev1.PodSpec{
					Affinity:           affinity,
					Tolerations:        tolerations,
					ServiceAccountName: "ugcupload-jmeter",
					Volumes: []corev1.Volume{
						{
							Name:         "telegraf-config-map",
							VolumeSource: volumeSource,
						},
						{
							Name:         "test-output-dir",
							VolumeSource: testOuputVolumeSource,
						},
					},
					Containers: []corev1.Container{
						{
							TTY:   true,
							Stdin: true,
							Name:  "jmslave",
							Image: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/ugctestgrid/jmeter-slave:latest", strconv.FormatInt(awsAcntNbr, 10), awsRegion),
							Args:  []string{"/bin/bash", "-c", "--", "/fileupload/upload > /fileupload.log 2>&1"},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: int32(1099)},
								corev1.ContainerPort{ContainerPort: int32(50000)},
								corev1.ContainerPort{ContainerPort: int32(1007)},
								corev1.ContainerPort{ContainerPort: int32(5005)},
								corev1.ContainerPort{ContainerPort: int32(8778)},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-output-dir",
									MountPath: "/test-output",
								},
							},
							Resources: resourcerequirementSlave,
						},
						{
							Name:  "telegraf",
							Image: "docker.io/telegraf:1.11.5-alpine",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: int32(8125)},
								corev1.ContainerPort{ContainerPort: int32(8092)},
								corev1.ContainerPort{ContainerPort: int32(8094)},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "telegraf-config-map",
									MountPath: "/etc/telegraf/telegraf.conf",
									SubPath:   "telegraf.conf",
								},
								{
									Name:      "test-output-dir",
									MountPath: "/test-output",
								},
							},
							Resources: resourcerequirements,
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment for slave...")
	result, e := deploymentsClient.Create(deployment)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems creating deployment for slave")
		created = false
		err = e.Error()
	} else {
		log.WithFields(log.Fields{
			"name": result.GetObjectMeta().GetName(),
		}).Info("Deployment succesful created deployment for slave(s")
		created = true
	}

	return
}

//CreateJmeterSlaveService creates service for jmeter slave
func (kop *Operations) CreateJmeterSlaveService(ns string) (created bool, err string) {

	res, e := kop.ClientSet.CoreV1().Services(ns).Create(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jmeter-slaves-svc",
			Namespace: ns,
			Labels: map[string]string{
				"jmeter_mode": "slave",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{Name: "first", Port: int32(1099), TargetPort: intstr.IntOrString{StrVal: "1099"}},
				corev1.ServicePort{Name: "second", Port: int32(5000), TargetPort: intstr.IntOrString{StrVal: "5000"}},
				corev1.ServicePort{Name: "fileupload", Port: int32(1007), TargetPort: intstr.IntOrString{StrVal: "1007"}},
				corev1.ServicePort{Name: "jolokia", Port: int32(8778), TargetPort: intstr.IntOrString{StrVal: "8778"}},
			},
			Selector: map[string]string{
				"jmeter_mode": "slave",
			},
		},
	})

	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems creating service for slave")
		created = false
		err = e.Error()
	} else {
		log.WithFields(log.Fields{
			"name": res.GetObjectMeta().GetName(),
		}).Info("Deployment succesful created service for slave")
		created = true
	}

	return

}

//CreateJmeterMasterDeployment used to create jmeter master deployment
func (kop *Operations) CreateJmeterMasterDeployment(namespace string, awsAcntNbr int64, awsRegion string) (created bool, err string) {

	deploymentsClient := kop.ClientSet.AppsV1().Deployments(namespace)
	values := []string{"master"}
	nodeSelectorRequirement := corev1.NodeSelectorRequirement{Key: "jmeter_mode", Operator: corev1.NodeSelectorOpIn, Values: values}
	nodeSelectorRequirements := []corev1.NodeSelectorRequirement{nodeSelectorRequirement}
	nodeSelectorTerm := corev1.NodeSelectorTerm{MatchExpressions: nodeSelectorRequirements}
	nodeSelectorTerms := []corev1.NodeSelectorTerm{nodeSelectorTerm}
	nodeSelector := &corev1.NodeSelector{NodeSelectorTerms: nodeSelectorTerms}
	nodeAffinity := &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: nodeSelector}
	affinity := &corev1.Affinity{NodeAffinity: nodeAffinity}

	toleration := corev1.Toleration{Key: "jmeter_master", Operator: corev1.TolerationOpExists, Value: "", Effect: corev1.TaintEffectNoSchedule}
	tolerations := []corev1.Toleration{toleration}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "jmeter-master",
			Labels: map[string]string{
				"jmeter_mode": "master",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"jmeter_mode": "master",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jmeter_mode": "master",
					},
				},
				Spec: corev1.PodSpec{
					Affinity:           affinity,
					Tolerations:        tolerations,
					ServiceAccountName: "ugcupload-jmeter",
					Containers: []corev1.Container{
						{
							TTY:   true,
							Stdin: true,
							Name:  "jmmaster",
							Image: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/ugctestgrid/jmeter-master:latest", strconv.FormatInt(awsAcntNbr, 10), awsRegion),
							Args:  []string{"/bin/bash", "-c", "--", "while true; do sleep 30; done;"},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:  int64Ptr(1000),
								RunAsGroup: int64Ptr(1000),
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: int32(60000)},
								corev1.ContainerPort{ContainerPort: int32(1025)},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	result, e := deploymentsClient.Create(deployment)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems creating deployment")
		created = false
		err = e.Error()
	} else {
		log.WithFields(log.Fields{
			"name": result.GetObjectMeta().GetName(),
		}).Info("Deployment succesful")
		created = true
	}
	return
}

//GetPodIpsForSlaves used to get the endpoints associated with a service
func (kop *Operations) GetPodIpsForSlaves(ns string) (endpoints []string) {
	var eps []string
	ep, e := kop.ClientSet.CoreV1().Endpoints(ns).Get("jmeter-slaves-svc", metav1.GetOptions{})
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
		}).Error("Problems getting endpoint for the service")
	} else {

		for _, epsub := range ep.Subsets {
			for _, epa := range epsub.Addresses {
				log.WithFields(log.Fields{
					"IP":       epa.IP,
					"Hostname": epa.Hostname,
				}).Info("Endpoint address")
				eps = append(eps, string(epa.IP))
			}
		}
	}
	endpoints = eps
	return
}

//GetHostNamesOfJmeterMaster Gets the ip addresses of the master
func (kop *Operations) GetHostNamesOfJmeterMaster(ns string) (hostnames []string) {

	var hn []string
	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"jmeter_mode": "master"}}
	actual := metav1.ListOptions{LabelSelector: labels.Set(labelSelector.MatchLabels).String()}
	pods, err := kop.ClientSet.CoreV1().Pods(ns).List(actual)
	if err != nil {
		log.WithFields(log.Fields{
			"err":       err.Error(),
			"namespace": ns,
		}).Error("Unable to find any pods in the namespace")
	} else {

		for _, pod := range pods.Items {
			log.WithFields(log.Fields{
				"hostIP": pod.Status.PodIP,
				"name":   pod.Name,
			}).Info("Jmeter slaves")
			if strings.EqualFold(string(pod.Status.Phase), "Running") {
				hn = append(hn, pod.Status.PodIP)
			}
		}
		hostnames = hn
	}
	return
}

//CheckNamespaces check for the existence of a namespace
func (kop *Operations) CheckNamespaces(namespace string) (exist bool) {
	var list corev1.NamespaceList
	d, err := kop.ClientSet.RESTClient().Get().AbsPath("/api/v1/namespaces").Param("pretty", "true").DoRaw()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Unable to retrieve all namespaces")
	} else {
		if err := json.Unmarshal(d, &list); err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("unmarsll the namespaces response")
		}

		exist = false
		for _, ns := range list.Items {
			if ns.Name == namespace {
				log.WithFields(log.Fields{
					"namespace": ns.Name,
				}).Info("name spaces found")
				exist = true
			}

		}
	}
	return
}

//LoadBalancerIP gets the loadbalancer ip of the service
func (kop *Operations) LoadBalancerIP(ns string, svc string) string {
	services, err := kop.ClientSet.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err.Error(),
			"service": svc,
		}).Errorf("Unable to get service: %s", ns)
	} else {

		log.WithFields(log.Fields{
			"number of services": len(services.Items),
			"namespace":          ns,
			"svc":                svc,
		}).Info("Number of service")
		for _, service := range services.Items {

			var ips []string
			for _, ing := range service.Status.LoadBalancer.Ingress {
				ips = append(ips, ing.IP)
			}
			var hostname []string
			for _, ing := range service.Status.LoadBalancer.Ingress {
				ips = append(ips, ing.Hostname)
			}

			if strings.ToLower(service.Name) == strings.ToLower(svc) {
				log.WithFields(log.Fields{
					"service.Name":                service.Name,
					"svc":                         svc,
					"service.Spec.LoadBalancerIp": service.Spec.LoadBalancerIP,
					"service.Spec.ExeternalIPs":   strings.Join(service.Spec.ExternalIPs, ","),
					"service.Status.LoadBalancer.Ingress.Hostname": hostname,
					"service.Status.LoadBalancer.Ingress.IP":       strings.Join(ips, ""),
				}).Info("Service Details")
				fmt.Println(fmt.Sprintf("------ ipe[0]", strings.Join(ips, "")))
				return strings.Join(ips, "")
			}
		}
	}
	return ""
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

//RegisterClient used to register the client
func (kop *Operations) RegisterClient() (success bool) {
	// creates the clientset
	kop.Init()
	clientset, err := kubernetes.NewForConfig(kop.Config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to register client")
		success = false
	} else {
		kop.ClientSet = clientset
		success = true
	}
	return
}

//GenerateReport creates report for tenant
func (kop Operations) GenerateReport(data string) (created bool, err string) {

	se := shellExec.Exec{}
	args := []string{data}
	_, err = se.ExecuteCommand("gen-report.py", args)
	if err != "" {
		log.WithFields(log.Fields{
			"err":  err,
			"data": data,
			"args": strings.Join(args, ","),
		}).Errorf("unable to generate the report")
		created = false
	} else {
		created = true
	}
	return
}

//CreateServiceaccount create service account
func (kop Operations) CreateServiceaccount(ns string, policyarn string) (created bool, err string) {

	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "create-serviceaccount.sh")
	args := []string{ns, policyarn}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("unable to create the service account in workspace: %v", ns)
		created = false
	} else {
		created = true
	}
	return
}

//DeleteServiceAccount deletes the service account
func (kop Operations) DeleteServiceAccount(ns string) (deleted bool, err string) {

	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "delete-serviceaccount.sh")
	args := []string{ns}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("unabme able to delete the service account in workspace: %v", ns)
		deleted = false
	} else {
		deleted = true
	}
	return
}

//StopTest stops the test in the namespace
func (kop Operations) StopTest(ns string) (started bool, err string) {
	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "stop_test.sh")
	args := []string{ns}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("unable to stop the test %v", strings.Join(args, ","))
		started = false
	} else {
		started = true
	}
	return
}
