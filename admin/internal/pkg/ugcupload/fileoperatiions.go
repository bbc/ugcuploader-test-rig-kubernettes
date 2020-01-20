package ugcupload

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"

	"bytes"

	uuid "github.com/satori/go.uuid"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//FileUploadOperations used to perform file upload
type FileUploadOperations struct {
	Context *gin.Context
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

func (fop FileUploadOperations) UploadFile(file io.Reader, uri string, destFileName string) {

	//prepare the reader instances to encode
	extraParams := map[string]string{
		"name": destFileName,
	}

	request, err := newfileUploadRequest(file, uri, extraParams, destFileName)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		var bodyContent []byte
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		resp.Body.Read(bodyContent)
		resp.Body.Close()
		fmt.Println(bodyContent)
	}
}

//ProcessData used to copy the supplied data file to right location
func (fop FileUploadOperations) ProcessData(uri string) (destFilename string) {

	file, err := fop.Context.FormFile("data")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Unable to get the test data from the form")
	}

	if file != nil {
		log.Println(file.Filename)
		f, err := file.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"err":      err.Error(),
				"filename": file.Filename,
			}).Error("Could not open the file")
		} else {
			fop.UploadFile(f, uri, file.Filename)
		}
		f.Close()
		//fop.Context.SaveUploadedFile(file, props.MustGet("data")+"/"+file.Filename)
	}
	return
}

//UploadJmeterProps use to upload the jmeter property file
func (fop FileUploadOperations) UploadJmeterProps(uri string, bw string) {

	home := os.Getenv("HOME")
	bwLock := fmt.Sprintf("%s/config/bandwidth/%s/bandwidth.csv", home, bw)

	r, err := os.Open(bwLock)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"filename": bwLock,
			"ur":       uri,
		}).Error("Could not open bandwidth file")
	}
	fop.UploadFile(r, uri, "jmeter.properties")
	r.Close()

}

//ProcessJmeter used to copy the supplied jmeter file to the right lcoation
func (fop FileUploadOperations) ProcessJmeter() (testFile string) {

	t := time.Now()
	u2 := fmt.Sprintf("%s-%s", uuid.NewV4(), t.Format("20060102150405"))
	path := fmt.Sprintf("%s/%s", props.MustGet("jmeter"), u2)
	fmt.Println(path)
	os.MkdirAll(path, os.ModePerm)
	jmeterScript, err := fop.Context.FormFile("jmeter")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to get the jmeter script from the form")
		return
	}

	if jmeterScript != nil {
		destFileName := fmt.Sprintf("%s/%s", path, jmeterScript.Filename)
		fop.Context.SaveUploadedFile(jmeterScript, destFileName)
		testFile = fmt.Sprintf("%s/%s", u2, jmeterScript.Filename)
	}

	return
}
