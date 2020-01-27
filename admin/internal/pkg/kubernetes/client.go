package kubernetes

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/magiconair/properties"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	shellExec "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/exec"
	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
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

//DeleteDeployment used to delete a deployment
func (kop *Operations) DeleteDeployment(namespace string) (deleted bool) {
	// Delete Deployment
	deletePolicy := metav1.DeletePropagationForeground
	deploymentsClient := kop.ClientSet.AppsV1().Deployments(namespace)
	if err := deploymentsClient.Delete(namespace, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Problem deleting deployment: %s", namespace)
		deleted = false
	} else {
		deleted = true
	}
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

//GetAllNodesWithPods gets all pods within a namespace
func (kop *Operations) GetAllNodesWithPods(ns string) (pds []string, found bool) {
	var pods []string
	actual := metav1.ListOptions{}
	res, e := kop.ClientSet.CoreV1().Pods(ns).List(actual)
	if e != nil {
		log.WithFields(log.Fields{
			"err":       e.Error(),
			"namespace": ns,
		}).Error("Problems getting all pods in namespace")
		found = false
	} else {
		for _, item := range res.Items {
			pods = append(pods, item.Spec.NodeName)
		}
		pds = pods
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

//CreateJmeterSlaveDeployment creates deployment for jmeter slaves
func (kop *Operations) CreateJmeterSlaveDeployment(ns string, nbrnodes int32, awsAcntNbr int64, awsRegion string) (created bool, err string) {

	deploymentsClient := kop.ClientSet.AppsV1().Deployments(ns)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "jmeter-slave",
			Labels: map[string]string{
				"jmeter_mode": "slave",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(nbrnodes),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"jmeter_mode": "slave",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jmeter_mode": "slave",
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"jmeter_mode": "slaves",
					},
					ServiceAccountName: "ugcupload-jmeter",
					Containers: []corev1.Container{
						{
							TTY:   true,
							Stdin: true,
							Name:  "jmslave",
							Image: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/ugctestgrid/jmeter-slave:latest", strconv.FormatInt(awsAcntNbr, 10), awsRegion),
							Args:  []string{"/bin/bash", "-c", "--", "/fileupload/upload > /fileuplouad.log 2>&1"},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{ContainerPort: int32(1099)},
								corev1.ContainerPort{ContainerPort: int32(50000)},
								corev1.ContainerPort{ContainerPort: int32(1007)},
							},
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
					NodeSelector: map[string]string{
						"jmeter_mode": "master",
					},
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

//GetPodIpsForSlaves used to get the endpoints assoicated with a service
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

//GetHostNamesOfJmeterMaster Gets the ip addresses of the slaves
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

		//<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>
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

//GetHostNamesOfJmeterSlaves Gets the ip addresses of the slaves
func (kop *Operations) GetHostNamesOfJmeterSlaves(ns string) (hostnames []string) {

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

		//<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>
		for _, pod := range pods.Items {
			log.WithFields(log.Fields{
				"hostIP": pod.Status.HostIP,
				"name":   pod.Name,
			}).Info("Jmeter slaves")
			if strings.EqualFold(string(pod.Status.Phase), "Running") {
				hn = append(hn, pod.Status.HostIP)
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

//LoadBalancerIP gets the load balancer ip of the service
func (kop *Operations) LoadBalancerIP(namespace string) (host string) {

	var list corev1.ServiceList
	err := kop.ClientSet.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/namespaces/%s/services", namespace)).Param("pretty", "true").Do().Into(&list)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to get service: %s", namespace)
	} else {
		for _, svc := range list.Items {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				host = ingress.Hostname
			}
		}
	}
	return
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

//CheckIfRunningJava Used to check if the pod has java running
func (kop Operations) CheckIfRunningJava(ns string, pod string) (resp string, err string) {
	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "check-if-jmeter-running.sh")
	args := []string{ns, pod}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err":       err,
			"namespace": ns,
			"pod":       pod,
		}).Errorf("failed checking if java is running on the pod")
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

//StartTest starts the uploaded test
func (kop Operations) StartTest(testPath string, ns string, listOfNodes string) (started bool, err string) {
	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "start_test_controller.sh")
	args := []string{testPath, ns, listOfNodes}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)

	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("unable to start the test %v", strings.Join(args, ","))
		started = false
	} else {
		started = true
	}
	return
}

//WaitForPodsToStart polls all pods in namespace until they are all running
func (kop Operations) WaitForPodsToStart(ns string, pods int) (started bool, err string) {
	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "wait-for-deployments.sh")
	args := []string{ns, strconv.Itoa(pods)}
	se := shellExec.Exec{}
	_, err = se.ExecuteCommand(cmd, args)
	if err != "" {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorf("slave pods did not start %v", strings.Join(args, ","))
		started = false
	} else {
		started = true
	}
	return
}
