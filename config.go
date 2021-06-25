package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"strconv"
	"strings"
	"text/template"
)

const MaxUsers = 10000

type config struct {
	finalNumber     *int
	ccServer        *string
	ccUserName      *string
	ccPassword      *string
	axlServer       *string
	axlUserName     *string
	axlUserPassword *string
	timeOut         *int
}

var (
	Config = config{
		finalNumber:     kingpin.Flag("number", fmt.Sprintf("Number of expected users (0 - %d)", MaxUsers)).Short('n').PlaceHolder("X").Required().Int(),
		ccServer:        kingpin.Flag("uccx", "UCCX server FQDN or IP address").Short('x').PlaceHolder("FQDN").Required().String(),
		ccUserName:      kingpin.Flag("user", "UCCX administrator name").PlaceHolder("USER").Short('u').Required().String(),
		ccPassword:      kingpin.Flag("pwd", "UCCX administrator password").PlaceHolder("PWD").Short('p').Required().String(),
		axlServer:       kingpin.Flag("cucm", "CUCM Publisher FQDN or IP address").PlaceHolder("FQD").Short('c').Required().String(),
		axlUserName:     kingpin.Flag("axl", "CUCM AXL administrator name").Short('a').PlaceHolder("USER").Required().String(),
		axlUserPassword: kingpin.Flag("axlpwd", "CUCM AXL administrator password").PlaceHolder("PWD").Short('s').Required().String(),
		timeOut: kingpin.Flag("timeout", fmt.Sprintf("Request timeout in seconds (%d - %d", MinTimeout, MAxTimeout)).
			PlaceHolder(fmt.Sprintf("%d", DefaultTimeout)).Short('t').Default("30").Int(),
	}
	logLevel = kingpin.Flag("level", "Logging level (Fatal, Error, Warning, Info, Debug, Trace)").Short('l').PlaceHolder("INFO").Default("Info").String()
)

func (c *config) validate() bool {
	c.isValidTimeOut()
	err := c.isValidNumbers()
	success := true
	if err != nil {
		log.Error(err)
		success = false
	}
	err = c.isValidCcxServer()
	if err != nil {
		log.Error(err)
		success = false
	}
	err = c.isValidCucmServer()
	if err != nil {
		log.Error(err)
		success = false
	}
	return success

}

func (c *config) isValidNumbers() error {
	log.Tracef("test if [%d] is in valid range", *c.finalNumber)
	if *c.finalNumber < 0 || *c.finalNumber > MaxUsers {
		return fmt.Errorf("expected number between 0 and %d", MaxUsers)
	}
	return nil
}

func (c *config) isValidTimeOut() {
	log.Tracef("test if [%d] is in range 5 - 120", *c.timeOut)
	if *c.timeOut < 5 || *c.timeOut > 120 {
		*c.timeOut = DefaultTimeout
	}
}

func (c *config) isValidCucmServer() error {
	return isValidServer(*c.axlServer, "CUCM")
}

func (c *config) isValidCcxServer() error {
	return isValidServer(*c.ccServer, "CCX")
}

func getLevel() log.Level {
	switch strings.ToLower(*logLevel) {
	case "t", "tra", "trace":
		return log.TraceLevel
	case "d", "deb", "debug":
		return log.DebugLevel
	case "i", "inf", "info":
		return log.InfoLevel
	case "w", "war", "warn", "warning":
		return log.WarnLevel
	case "e", "err", "error":
		return log.ErrorLevel
	case "f", "fat", "fatal":
		return log.FatalLevel
	default:
		return DefaultLogLevel
	}
}

func (c *config) toString() string {
	infoTmpl = `
Actual run configuration:
  Max users       {{.number}}
  CCX Server      {{.ccx}}
  CCX Admin       {{.ccxUser}}
  CUCM Server     {{.cucm}}
  CUCM Admin      {{.cucmUser}}
  Timeout         {{.timeOut}}
  Log level       {{.level}}
`

	usr := strconv.Itoa(*c.finalNumber)
	if *c.finalNumber == 0 {
		usr = "delete all generated"
	}

	var m = map[string]string{
		"number":   usr,
		"ccx":      *c.ccServer,
		"ccxUser":  *c.ccUserName,
		"cucm":     *c.axlServer,
		"cucmUser": *c.axlUserName,
		"timeOut":  strconv.Itoa(*c.timeOut),
		"level":    *logLevel,
	}
	t := template.Must(template.New("info").Parse(infoTmpl))
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "info", m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}
