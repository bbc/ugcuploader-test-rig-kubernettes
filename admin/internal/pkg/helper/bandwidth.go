package helper

import (
	mytype "github.com/bbc/ugcuploader-test-rig-kubernettes/admin/internal/pkg/types"
)

type Bandwidth struct {
	BandwithSelection map[string]mytype.BandwidthSelection
}
func (bw *Bandwidth) Init(){
	bws := make(map[string]mytype.BandwidthSelection)
	bws["adsl"] = mytype.BandwidthSelection{
			HTTPSCPS: 1024000,
			HTTPCPS: 1024000,
			TIMEOUT: 3600000}
	bws["adsl2"] = mytype.BandwidthSelection{
		HTTPSCPS: 1536000,
		HTTPCPS: 1536000,
		TIMEOUT: 3600000}
	bws["adsl2Plus"]  = mytype.BandwidthSelection{
		HTTPSCPS: 3072000,
		HTTPCPS: 3072000,
		TIMEOUT: 3600000}
	bws["ethernetLan"]  = mytype.BandwidthSelection{
		HTTPSCPS: 1280000,
		HTTPCPS: 1280000,
		TIMEOUT: 3600000}
	bws["fastEthernet"]  = mytype.BandwidthSelection{
		HTTPSCPS: 12800000,
		HTTPCPS: 12800000,
		TIMEOUT: 3600000}
	bws["gigabitEthernet"]  = mytype.BandwidthSelection{
		HTTPSCPS: 128000000,
		HTTPCPS: 128000000,
		TIMEOUT: 3600000}
	bws["10gigabitEthernet"]  = mytype.BandwidthSelection{
		HTTPSCPS: 1280000000,
		HTTPCPS: 1280000000,
		TIMEOUT: 3600000}
	bws["100gigabitEthernet"]  = mytype.BandwidthSelection{
		HTTPSCPS: 12800000000,
		HTTPCPS: 12800000000,
		TIMEOUT: 3600000}
	bws["mobileDataEdge"]  = mytype.BandwidthSelection{
		HTTPSCPS: 49152,
		HTTPCPS: 49152,
		TIMEOUT: 3600000}
	bws["mobileDataHspa"]  = mytype.BandwidthSelection{
		HTTPSCPS: 1843200,
		HTTPCPS: 1843200,
		TIMEOUT: 3600000}
	bws["mobileDatacHspaPlus"]  = mytype.BandwidthSelection{
		HTTPSCPS: 2688000,
		HTTPCPS: 2688000,
		TIMEOUT: 3600000}
	bws["mobileDataDcHspaPlus"]  = mytype.BandwidthSelection{
		HTTPSCPS: 5376000,
		HTTPCPS: 5376000,
		TIMEOUT: 3600000}
	bws["mobileDataLte"]  = mytype.BandwidthSelection{
		HTTPSCPS: 19200000,
		HTTPCPS: 19200000,
		TIMEOUT: 3600000}
	bws["mobileDataGprs"]  = mytype.BandwidthSelection{
		HTTPSCPS: 21888,
		HTTPCPS: 21888,
		TIMEOUT: 3600000}
	bws["wifi80211a"]  = mytype.BandwidthSelection{
		HTTPSCPS: 6912000,
		HTTPCPS: 6912000,
		TIMEOUT: 3600000}
	bws["wifi80211n"]  = mytype.BandwidthSelection{
		HTTPSCPS: 76800000,
		HTTPCPS: 76800000,
		TIMEOUT: 3600000}

	bw.BandwithSelection = bws
}

func (bw Bandwidth) GetBandwidth(bandwidth string) (mytype.BandwidthSelection) {

	return bw.BandwithSelection[bandwidth]
}
