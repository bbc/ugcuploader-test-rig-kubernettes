package ugcupload

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//FileUploadOperations used to perform file upload
type FileUploadOperations struct {
	Form    *multipart.Form
	Context *gin.Context
}

//SaveFile used to copy the supplied data file to right location
func (fop FileUploadOperations) SaveFile(fileName string) {

	file, err := fop.Context.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Errorf("Unable to get the test data from the form")
	}

	if file != nil {
		log.Println(file.Filename)
		fop.Context.SaveUploadedFile(file, fileName)
	}

}
