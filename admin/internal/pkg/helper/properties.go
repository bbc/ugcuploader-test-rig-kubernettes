package helper

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"text/template"
)

type JmeterProperties struct{}

func (jp *JmeterProperties) Create(bandwidth string) string {

	bw := Bandwidth{}
	bw.Init()
	bwd := bw.GetBandwidth(bandwidth)

	home := os.Getenv("HOME")
	bwLock := fmt.Sprintf("%s/config/jmeter.properties.tmpl", home)
	t, err := template.New("jmeter.properties.tmpl").ParseFiles(bwLock)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"filename": bwLock,
		}).Error("Problems parsing template")
	}

	tmpfile, err := ioutil.TempFile(os.TempDir(), "jmeter-properties")
	defer tmpfile.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"filename": bwLock,
		}).Error("Could not create temporary file")
	}

	err = t.Execute(tmpfile, bwd)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err.Error(),
			"filename": bwLock,
		}).Error("Problems applying the template")
	} else {
		return tmpfile.Name()
	}

	return ""
}
