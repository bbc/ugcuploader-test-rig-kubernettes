package kubernetes

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/magiconair/properties"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
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

//Tenant Information about the tenant
type Tenant struct {
	Name      string
	Namespace string
	Running   bool
}

//GetallTenants Retuns a list of tenants
func (kop *Operations) GetallTenants() (ts []Tenant, err string) {
	tenants := []Tenant{}
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
			tenants = append(tenants, Tenant{Name: item.Name, Namespace: item.Namespace})
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
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jmeter_mode": "slave",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "ugcupload-jmeter",
					Containers: []v1.Container{
						{
							TTY:   true,
							Stdin: true,
							Name:  "jmmaster",
							Image: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/ugctestgrid/jmeter-slave:latest", strconv.FormatInt(awsAcntNbr, 10), awsRegion),
							Args:  []string{"/bin/bash", "-c", "--", "while true; do sleep 30; done;"},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{ContainerPort: int32(1099)},
								v1.ContainerPort{ContainerPort: int32(50000)},
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
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"jmeter_mode": "master",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "ugcupload-jmeter",
					Containers: []v1.Container{
						{
							TTY:   true,
							Stdin: true,
							Name:  "jmmaster",
							Image: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/ugctestgrid/jmeter-master:latest", strconv.FormatInt(awsAcntNbr, 10), awsRegion),
							Args:  []string{"/bin/bash", "-c", "--", "while true; do sleep 30; done;"},
							SecurityContext: &v1.SecurityContext{
								RunAsUser:  int64Ptr(1000),
								RunAsGroup: int64Ptr(1000),
							},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 60000,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	result, e := deploymentsClient.Create(deployment)
	if err != "" {
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

//CheckNamespaces check for the existence of a namespace
func (kop *Operations) CheckNamespaces(namespace string) (exist bool) {
	var list v1.NamespaceList
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

	var list v1.ServiceList
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

// Code below taken from here: https://github.com/kjk/go-cookbook/blob/master/advanced-exec/03-live-progress-and-capture-v2.go
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

//GenerateReport creates report for tenant
func (kop Operations) GenerateReport(data string) (created bool, err string) {

	args := []string{data}
	_, err = kop.executeCommand("gen-report.py", args)
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
	resp, err = kop.executeCommand(cmd, args)
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
	_, err = kop.executeCommand(cmd, args)
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
	_, err = kop.executeCommand(cmd, args)
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
	_, err = kop.executeCommand(cmd, args)
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
func (kop Operations) StartTest(testPath string, ns string, bandwidth string, nbrnodes int) (started bool, err string) {
	cmd := fmt.Sprintf("%s/%s", props.MustGet("tscripts"), "start_test_controller.sh")
	args := []string{testPath, ns, bandwidth, strconv.Itoa(nbrnodes)}
	//_, err = kop.executeCommand("start_test_controller.sh", args)
	_, err = kop.executeCommand(cmd, args)

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

func (kop Operations) executeCommand(command string, args []string) (outStr string, errStr string) {

	var logger = log.WithFields(log.Fields{
		"command": command,
		"args":    strings.Join(args, ","),
	})

	w := &logrusWriter{
		entry: logger,
	}

	cmd := exec.Command(command, args...)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	}

	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	stdout := NewCapturingPassThroughWriter(w)
	stderr := NewCapturingPassThroughWriter(w)
	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("unable to start the execute the command: %v", strings.Join(args, ","))
		errStr = err.Error()

	} else {

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
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Problems waiting for command to complete")
		}
		if errStdout != nil || errStderr != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("Error occured when logging the execution process")
		}
		os, te := string(stdout.Bytes()), string(stderr.Bytes())

		if te != "" && strings.Contains(te, "TTY - input is not a terminal") {
			log.WithFields(log.Fields{
				"err": te,
			}).Warn("TTY - input is not a terminal: %v", strings.Join(args, ","))
		} else {
			errStr = te
		}
		outStr = os

	}
	return

}

type logrusWriter struct {
	entry *log.Entry
	buf   bytes.Buffer
	mu    sync.Mutex
}

func (w *logrusWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	origLen := len(b)
	for {
		if len(b) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			w.buf.Write(b)
			return origLen, nil
		}

		w.buf.Write(b[:i])
		w.alwaysFlush()
		b = b[i+1:]
	}
}

func (w *logrusWriter) alwaysFlush() {
	w.entry.Info(w.buf.String())
	w.buf.Reset()
}

func (w *logrusWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.buf.Len() != 0 {
		w.alwaysFlush()
	}
}
