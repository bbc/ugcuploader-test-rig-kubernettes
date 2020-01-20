package jmeter

import (
	"os"

	"github.com/antchfx/xmlquery"
	log "github.com/sirupsen/logrus"
)

//Jmeter used to perform jmeter operations
type Jmeter struct {
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

			/jmeterTestPlan/hashTree/hashTree/hashTree/HTTPSamplerProxy/elementProp/collectionProp/elementProp

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
