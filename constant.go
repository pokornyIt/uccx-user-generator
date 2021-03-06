package main

import (
	log "github.com/sirupsen/logrus"
)

const (
	ApplicationName = "uccx-user-generator" // application name

	AxlUrlFormat = "https://%s:8443/axl/" // format string for SOAP URL
	AxlIdPrefix  = "axl-"                 // AXL SOAP request id start with

	// AxlXmlHeaderFormat AXL basic envelope strings
	AxlXmlHeaderFormat    = "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:ns=\"http://www.cisco.com/AXL/API/%s\">"
	AxlCcmVersion         = "<soapenv:Header/><soapenv:Body><ns:getCCMVersion sequence=\"%d\">\n</ns:getCCMVersion></soapenv:Body></soapenv:Envelope>"
	AxlSqlRequest         = "<soapenv:Header/><soapenv:Body><ns:executeSQLQuery sequence=\"%d\">\n<sql>%s</sql></ns:executeSQLQuery></soapenv:Body></soapenv:Envelope>"
	AxlRemoveCcxExtension = "<soapenv:Header/><soapenv:Body><ns:updateUser sequence=\"%d\">\n<uuid>%s</uuid>\n<ipccExtension /></ns:updateUser></soapenv:Body></soapenv:Envelope>"
	AxlSetCcxExtension    = "<soapenv:Header/><soapenv:Body><ns:updateUser sequence=\"%d\">\n<uuid>%s</uuid>\n<ipccExtension uuid=\"%s\">%s</ipccExtension></ns:updateUser></soapenv:Body></soapenv:Envelope>"
	AxlDbVersionError     = "Error"

	CcxUrlMainPart     = "/adminapi/"                     // CCX REST API path
	CcxUrlForce        = "/uccx-webservices/getAllAgents" // CCX force resource request
	CcxForceMaxUsers   = 50                               // Maximal user for effective CCX user force test 20,50,100
	CcxForceWaitTime   = 3                                // Seconds wait after CCX Force return data use 1,5,10
	CcxForceDownTime   = 5                                // minutes wait before repeat force update request
	CxxForceDownRepeat = 3                                // how many time system try force update
	CcxResourcePath    = "resource"                       // CCX resource part part
	CcxTeamPath        = "team"                           //CCX Team path part
	CcxIdPrefix        = "ccx-"                           // CCX REST API request id start with

	CcxTeamNameFormat = "Perf_test_%04d"    // format string for team
	CcxTeamNameRegex  = `^Perf_test_\d{4}$` // regex string for validate generated team
	CcxUserNameFormat = "perf_user_%06d"    // format string for user login name
	CcxUserNameRegex  = `^perf_user_\d{6}$` // regex string for validate generated user login name

	// TODO: help variable for c09 exist users
	_CcxUserNameRegex      = `^agent\d{2,3}$` // regex string for validate generated user login name c09 exist users
	CcxUserFirstNameFormat = "perf_user"      // user generated first name
	CcxUserLastNameFormat  = "perf_%06d"      // format string for generated user last name

	DefaultTimeout  = 30            // default request timeout in seconds
	MinTimeout      = 5             // minimal request timeout in seconds
	MAxTimeout      = 120           // maximal request timeout in seconds
	DefaultLogLevel = log.InfoLevel // default log level

	AxlRepeatRequestCount = 3 // number of repeat for read AXL API
	AxlRepeatWaitSeconds  = 5 // number of seconds wait before try read AXL api again

	TimeFormat         = "15:04:05.0000"           // time format
	DateTimeFormat     = "2006-01-02 15:04:05.000" // Full date time format
	DateTimeFileFormat = "20060102-150405"         // file date format
	StatFormatString   = "%s    %5s / %15s / %15s / %15s / %15s / %15s\r\n"
)

var (
	AxlDbVersions = []string{"10.0", "12.0", "14.0", "16.0", "18.0"} // list of supported AXL DB versions 10.x - 20.x
)
