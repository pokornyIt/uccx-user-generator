package main

import log "github.com/sirupsen/logrus"

const (
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

	CcxTeamNameFormat      = "Gen_test_%04d"        // format string for team
	CcxTeamNameRegex       = `^Gen_test_\d{4}$`     // regex string for validate generated team
	CcxUserNameFormat      = "gen_sso_user_%06d"    // format string for user login name
	CcxUserNameRegex       = `^gen_sso_user_\d{6}$` // regex string for validate generated user login name
	CcxUserFirstNameFormat = "sso_user"             // user generated first name
	CcxUserLastNameFormat  = "sso_%06d"             // format string for generated user last name

	DefaultTimeout  = 30            // default request timeout in seconds
	MinTimeout      = 5             // minimal request timeout in seconds
	MAxTimeout      = 120           // maximal request timeout in seconds
	DefaultLogLevel = log.InfoLevel // default log level
)

var (
	AxlDbVersions = []string{"10.0", "12.0", "14.0", "16.0", "18.0"} // list of supported AXL DB versions 10.x - 20.x
)
