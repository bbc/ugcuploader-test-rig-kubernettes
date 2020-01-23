package jmeter

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/antchfx/xmlquery"
	log "github.com/sirupsen/logrus"
)

//Jmeter used to perform jmeter operations
type Jmeter struct{}

//StopTestOnMaster Used to stop the test on master
func (jmeter Jmeter) StopTestOnMaster(uri string) (error string, res bool) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
			"uri": uri,
		}).Error("Unable to create the html request to stop the test")
		error = err.Error()
		res = false
		return
	}

	resp, errClient := client.Do(req)

	if errClient != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
			"uri": uri,
		}).Error("Problems making call to stop the test")
		res = false
		error = err.Error()
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	log.WithFields(log.Fields{
		"resp": string(body),
		"uri":  uri,
	}).Info("Successfully stop test")

	return

}

//StartTestOnMaster Used to start the test on master
func (jmeter Jmeter) StartTestOnMaster(testFile io.Reader, uri, tenant string, hosts string, testFileName string) (error string, res bool) {

	//prepare the reader instances to encode
	extraParams := map[string]string{
		"testfile": testFileName,
		"tenant":   tenant,
		"hosts":    hosts,
	}

	request, err := newfileUploadRequest(testFile, uri, extraParams, testFileName)
	if err != nil {
		log.Fatal(err)
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

	body, err := ioutil.ReadAll(resp.Body)
	log.WithFields(log.Fields{
		"err":         err.Error(),
		"statusCode":  resp.StatusCode,
		"bodyContent": string(body),
	}).Info("Response from starting jmeter tests")
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
