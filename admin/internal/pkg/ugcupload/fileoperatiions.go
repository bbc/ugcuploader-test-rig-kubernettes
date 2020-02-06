package ugcupload

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
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

func (fop FileUploadOperations) uploadFile(file io.Reader, uri string, destFileName string) (error string, uploaded bool) {

	//prepare the reader instances to encode
	extraParams := map[string]string{
		"name": destFileName,
	}

	request, err := newfileUploadRequest(file, uri, extraParams, destFileName)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("Problems creating request")
		error = err.Error()
		uploaded = false
	}

	client := &http.Client{}
	resp, errReq := client.Do(request)
	if errReq != nil {
		log.WithFields(log.Fields{
			"err": errReq.Error(),
		}).Error("Problems uploading file")
		error = errReq.Error()
		uploaded = false
		return
	}
	resp.Body.Close()
	uploaded = true
	return
}

//ProcessData used to copy the supplied data file to right location
func (fop FileUploadOperations) ProcessData(uri string, filename string, data *multipart.File) (error string, uploaded bool) {
	error, uploaded = fop.uploadFile(*data, uri, filename)
	return
}

//UploadJmeterProps use to upload the jmeter property file
func (fop FileUploadOperations) UploadJmeterProps(uri string, bw string) (error string, upload bool) {

	home := os.Getenv("HOME")
	bwLock := fmt.Sprintf("%s/config/bandwidth/%s/bandwidth.csv", home, bw)

	r, err := os.Open(bwLock)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"filename": bwLock,
			"ur":       uri,
		}).Error("Could not open bandwidth file")
		upload = false
		error = err.Error()
	}
	error, upload = fop.uploadFile(r, uri, "jmeter.properties")
	upload = true
	r.Close()
	return
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
