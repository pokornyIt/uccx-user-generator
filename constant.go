package main

import log "github.com/sirupsen/logrus"

const (
	AxlUrlFormat       = "https://%s:8443/axl"
	AxlIdPrefix        = "axl-"
	AxlXmlHeaderFormat = "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:ns=\"http://www.cisco.com/AXL/API/%s\">"
	AxlCcmVersion      = "<soapenv:Header/><soapenv:Body><ns:getCCMVersion sequence=\"%d\">\n</ns:getCCMVersion></soapenv:Body></soapenv:Envelope>"
	AxlSqlRequest      = "<soapenv:Header/><soapenv:Body><ns:executeSQLQuery sequence=\"%d\">\n<sql>%s</sql></ns:executeSQLQuery></soapenv:Body></soapenv:Envelope>"
	AxlDbVersionError  = "Error"

	CcxUrlMainPart = "/adminapi/"
	CcxIdPrefix    = "ccx-"

	DefaultTimeout  = 30
	MinTimeout      = 5
	MAxTimeout      = 120
	DefaultLogLevel = log.InfoLevel
)

var (
	AxlDbVersions = []string{"10.0", "12.0", "14.0", "16.0", "18.0"}
)
