package jmeter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/antchfx/xmlquery"
	redis "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/redis"
	types "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
	ugl "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/ugcupload"
	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//Jmeter used to perform jmeter operations
type Jmeter struct {
	Fop          ugl.FileUploadOperations
	Context      *gin.Context
	redisUtils   redis.Redis
	RedisTenant  types.RedisTenant
	Data         *multipart.File
	DataFilename string
	JmeterScript *multipart.File
}

//IsRunning check to see if jmeter is running on the pod
func (jmeter Jmeter) IsRunning(podIP string) (error string, res bool) {
	return jmeter.makeGetRequest(fmt.Sprintf("http://%s:1025/is-running", podIP))
}

//IsSlaveRunning check to see if jmeter is running on the pod
func (jmeter Jmeter) IsSlaveRunning(podIP string) (res bool) {

	resp, err := getRequest(fmt.Sprintf("http://%s:1007/is-running", podIP))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
			"uri": podIP,
		}).Error("Slave is not running")
		res = false
		return
	}

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"uri": podIP,
		}).Error("Slave is not running")
		res = false
		return
	}
	res = true
	return
}

func (jmeter Jmeter) sendToDataSlave(hn string, wg sync.WaitGroup, message sync.Map) {
	wg.Add(1)
	defer wg.Done()
	dataURI := fmt.Sprintf("http://%s:1007/data", hn)
	e, uploaded := jmeter.Fop.ProcessData(dataURI, jmeter.DataFilename, jmeter.Data)
	if uploaded == false {
		message.Store(fmt.Sprintf("ProblesmUploadingData", hn), e)
	}
}

func (jmeter Jmeter) sendPropertiesToSlave(ugcLoadRequest types.UgcLoadRequest, hn string, wg sync.WaitGroup, message sync.Map) {
	wg.Add(1)
	defer wg.Done()
	jmeterURI := fmt.Sprintf("http://%s:1007/jmeter-props", hn)
	e, uploaded := jmeter.Fop.UploadJmeterProps(jmeterURI, ugcLoadRequest.BandWidthSelection)
	if uploaded == false {
		message.Store(fmt.Sprintf("ProblesmUploadingJmeterProps@", hn), e)
	}
}

func (jmeter Jmeter) startSlave(ugcLoadRequest types.UgcLoadRequest, hn string, wg sync.WaitGroup, message *sync.Map) {
	wg.Add(1)
	defer wg.Done()
	params := map[string]string{
		"xmx":              ugcLoadRequest.Xmx,
		"xms":              ugcLoadRequest.Xms,
		"cpu":              ugcLoadRequest.CPU,
		"ram":              ugcLoadRequest.RAM,
		"maxMetaspaceSize": ugcLoadRequest.MaxMetaspaceSize,
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if err != nil {
		message.Store(fmt.Sprintf("%s-CreatingMultipartForm", hn), err.Error())
		return
	}

	r, _ := http.NewRequest("POST", fmt.Sprintf("http://%s:1007/start-server", hn), body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, errReq := client.Do(r)
	if errReq != nil {
		log.WithFields(log.Fields{
			"err": errReq.Error(),
			"url": fmt.Sprintf("http://%s:1007/start-server", hn),
		}).Error("Failed to start slave")
		message.Store(fmt.Sprintf("%s-StartingServer", hn), errReq.Error())
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message.Store(fmt.Sprintf("%s-ReadingResponse", hn), err.Error())
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	log.WithFields(log.Fields{
		"resp": bodyString,
		"hn":   hn,
	}).Info("Started Slave")

}

//SetupSlaves used to seutp the slaves for testing
func (jmeter Jmeter) SetupSlaves(ugcLoadRequest types.UgcLoadRequest, hostnames []string) (error string, result bool) {

	var setupSlaveWaitGroup sync.WaitGroup
	message := sync.Map{}
	for _, hn := range hostnames {
		go jmeter.sendToDataSlave(hn, setupSlaveWaitGroup, message)
		go jmeter.sendPropertiesToSlave(ugcLoadRequest, hn, setupSlaveWaitGroup, message)
	}
	setupSlaveWaitGroup.Wait()

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
		error = b.String()
		result = false
		return
	}

	var startSlaveWaitGroup sync.WaitGroup
	slaveStartMessage := new(sync.Map)
	for _, hn := range hostnames {
		go jmeter.startSlave(ugcLoadRequest, hn, startSlaveWaitGroup, slaveStartMessage)
	}
	startSlaveWaitGroup.Wait()

	responses = make(map[interface{}]interface{})
	slaveStartMessage.Range(func(k, v interface{}) bool {
		responses[k] = v
		return true
	})

	if len(responses) != 0 {
		b := new(bytes.Buffer)
		for key, value := range responses {
			fmt.Fprintf(b, "%s:\"%s\"\n", key, value)
		}
		error = b.String()
		result = false
		return
	}

	result = true
	return

}

func (jmeter Jmeter) unMarshallResponse(body io.ReadCloser) (resp types.JmeterResponse) {

	bdy, err := ioutil.ReadAll(body)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Problems Reading Reponse")
		return
	}
	var jr types.JmeterResponse
	errMarshall := json.Unmarshal(bdy, &jr)
	if errMarshall != nil {
		log.WithFields(log.Fields{
			"resp":  string(bdy),
			"error": errMarshall.Error(),
		}).Error("Unable to unmashall the response from jmeter pod")
		return
	}
	return jr
}

func getRequest(uri string) (res *http.Response, err error) {

	client := &http.Client{}

	req, e := http.NewRequest("GET", uri, nil)
	if e != nil {
		log.WithFields(log.Fields{
			"err": e.Error(),
			"uri": uri,
		}).Error("Unable to create the html request to stop the test")
		err = e
		return
	}

	return client.Do(req)
}

func (jmeter Jmeter) makeGetRequest(uri string) (error string, res bool) {

	resp, errClient := getRequest(uri)
	if errClient != nil {
		log.WithFields(log.Fields{
			"err": errClient.Error(),
			"uri": uri,
		}).Error("Problems making call to stop the test")
		res = false
		error = errClient.Error()
		return
	}

	var jmeterResponse = jmeter.unMarshallResponse(resp.Body)
	if jmeterResponse.Code != 200 {
		res = false
		error = jmeterResponse.Message
		return
	}

	log.WithFields(log.Fields{
		"resp": jmeterResponse,
		"uri":  uri,
	}).Info("GET Request to Jmeter Suceeded")

	res = true
	return
}

//StopTestOnMaster Used to stop the test on master
func (jmeter Jmeter) StopTestOnMaster(podIP string) (error string, res bool) {
	uri := fmt.Sprintf("http://%s:1025/stop-test", podIP)
	return jmeter.makeGetRequest(uri)
}

//KillMaster Used to stop the test on master
func (jmeter Jmeter) KillMaster(podIP string) (error string, res bool) {
	uri := fmt.Sprintf("http://%s:1025/kill", podIP)
	return jmeter.makeGetRequest(uri)
}

//KillSlave used to kill the slaves
func (jmeter Jmeter) KillSlave(podIP string) (error string, res bool) {
	uri := fmt.Sprintf("http://%s:1007/kill", podIP)
	return jmeter.makeGetRequest(uri)

}

//StartMasterTest used to start the master and tests
func (jmeter Jmeter) StartMasterTest(master string, ugcLoadRequest types.UgcLoadRequest, listOfHost string) (error string, started bool) {

	uri := fmt.Sprintf("http://%s:1025/start-test", master)
	t := time.Now()
	u2 := fmt.Sprintf("%s-%s", uuid.NewV4(), t.Format("20060102150405"))
	path := fmt.Sprintf("%s/%s", props.MustGet("jmeter"), u2)
	return jmeter.startTestOnMaster(*jmeter.JmeterScript, uri, ugcLoadRequest.Context, listOfHost, path)
}

//StartTestOnMaster Used to start the test on master
func (jmeter Jmeter) startTestOnMaster(testFile io.Reader, uri, tenant string, hosts string, testFileName string) (error string, res bool) {

	//prepare the reader instances to encode
	extraParams := map[string]string{
		"testfile": testFileName,
		"tenant":   tenant,
		"hosts":    hosts,
	}

	request, err := newfileUploadRequest(testFile, uri, extraParams, testFileName)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems creting request 3")
		error = err.Error()
		res = false
		return
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems starting jmeter tests on master")
		error = err.Error()
		res = false
		return
	}

	var jmeterResponse = jmeter.unMarshallResponse(resp.Body)
	if jmeterResponse.Code != 200 {
		log.WithFields(log.Fields{
			"err":        jmeterResponse.Message,
			"statusCode": jmeterResponse.Code,
		}).Error("Response from starting jmeter tests")
		error = jmeterResponse.Message
		res = false
		return
	}
	res = true
	return
}

//GetFileName used to get the filname from the jmeter script
func (jmeter Jmeter) GetFileName(fn string) {

	f, err := os.Open(fn)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to open the jmeter script")
	} else {
		doc, err := xmlquery.Parse(f)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Errorf("Unable to initialize kubeconfig")
		} else {

			list := xmlquery.Find(doc, "//TestPlan[HTTPSamplerProxy[@enabled='true']/elementProp/collectionProp/elementProp/stringProp[@name='File.Path']")
			for _, l := range list {
				log.WithFields(log.Fields{
					"item": l,
				}).Info("Item from jmeter script")
			}

		}

		f.Close()
	}
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(file io.Reader, uri string, params map[string]string, filename string) (r *http.Request, e error) {

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	r, e = http.NewRequest("POST", uri, body)

	r.Header.Set("Content-Type", writer.FormDataContentType())
	return

}
