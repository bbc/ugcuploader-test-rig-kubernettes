package ugcupload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"time"

	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
)

var props = properties.MustLoadFile("/etc/ugcupload/loadtest.conf", properties.UTF8)

//FileUploadOperations used to perform file upload
type FileUploadOperations struct {
	Form *multipart.Form
}

//ProcessData used to copy the supplied data file to right location
func (fop FileUploadOperations) ProcessData() (destFilename string) {

	dataFiles := fop.Form.File["data"]
	for _, f := range dataFiles {
		src, err := f.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Errorf("Unable to open file at :%s", f.Filename)
		} else {
			defer src.Close()
			// Destination
			data := props.MustGet("data")
			destFilename = fmt.Sprintf("%s/%s", data, f.Filename)
			dst, err := os.Create(destFilename)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Errorf("Unable to create file at :%s", destFilename)
			} else {
				defer dst.Close()
				if _, err = io.Copy(dst, src); err != nil {
					log.WithFields(log.Fields{
						"err": err.Error(),
					}).Error("Unable to copy file")
				}
			}
		}

	}
	return
}

//ProcessJmeter used to copy the supplied jmeter file to the right lcoation
func (fop FileUploadOperations) ProcessJmeter() (testFile string) {

	t := time.Now()
	u2 := fmt.Sprintf("%s-%s", uuid.NewV4(), t.Format("20060102150405"))
	dataFiles := fop.Form.File["jmeter"]
	path := fmt.Sprintf("%s/%s", props.MustGet("jmeter"), u2)
	fmt.Println(path)
	os.MkdirAll(path, os.ModePerm)
	for _, f := range dataFiles {
		src, err := f.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Errorf("Unable to open file at :%s", path)
		} else {
			defer src.Close()
			destFileName := fmt.Sprintf("%s/%s", path, f.Filename)
			dst, err := os.Create(destFileName)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err.Error(),
				}).Errorf("Unable to create the destination file :%s", destFileName)
			} else {
				defer dst.Close()
				if _, err = io.Copy(dst, src); err != nil {
					log.WithFields(log.Fields{
						"err": err.Error(),
					}).Error("Unbale to copy to the destination")
				}
				testFile = fmt.Sprintf("%s/%s", u2, f.Filename)
			}
		}

	}
	return
}
