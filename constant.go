package main

import (
	log "github.com/sirupsen/logrus"
)

const (
	ApplicationName = "uccx-user-generator" // application name

	AxlUrlFormat = "https://%s:8443/axl/" // format string for SOAP URL
	AxlIdPrefix  = "axl-"                 // AXL SOAP request id start with

	// AxlXmlHeaderFormat AXL basic envelope strings
	AxlXmlHeaderFormat = "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:ns=\"http://www.cisco.com/AXL/API/%s\">"
	AxlCcmVersion      = "<soapenv:Header/><soapenv:Body><ns:getCCMVersion sequence=\"%d\">\n</ns:getCCMVersion></soapenv:Body></soapenv:Envelope>"
	AxlSqlRequest      = "<soapenv:Header/><soapenv:Body><ns:executeSQLQuery sequence=\"%d\">\n<sql>%s</sql></ns:executeSQLQuery></soapenv:Body></soapenv:Envelope>"
	AxlDbVersionError  = "Error"

	CcxUrlMainPart  = "/adminapi/" // CCX REST API path
	CcxResourcePath = "resource"   // CCX resource part part
	CcxTeamPath     = "team"       //CCX Team path part
	CcxIdPrefix     = "ccx-"       // CCX REST API request id start with

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

	TimeFormat     = "15:04:05.0000"           // time format
	DateTimeFormat = "2006-01-02 15:04:05.000" // Full date time format

)

var (
	AxlDbVersions = []string{"10.0", "12.0", "14.0", "16.0", "18.0"} // list of supported AXL DB versions 10.x - 20.x
)
