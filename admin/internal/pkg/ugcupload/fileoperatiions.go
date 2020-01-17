package ugcupload

import (
	"fmt"
	"mime/multipart"
	"os"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//FileUploadOperations used to perform file upload
type FileUploadOperations struct {
	Form    *multipart.Form
	Context *gin.Context
}

//ProcessData used to copy the supplied data file to right location
func (fop FileUploadOperations) ProcessData() (destFilename string) {

	file, err := fop.Context.FormFile("data")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to get the test data from the form")
	}

	if file != nil {
		log.Println(file.Filename)
		fop.Context.SaveUploadedFile(file, props.MustGet("data")+"/"+file.Filename)
	}
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
