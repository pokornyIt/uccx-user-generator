package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type AxlResponse struct {
	id            string
	url           string
	response      *http.Response
	err           error
	lastMessage   string
	body          string
	statusCode    int
	statusMessage string
}

func (a *AxlResponse) Close() {
	if a.response != nil && a.response.Body != nil {
		_ = a.response.Body.Close()
	}
	a.response = nil
}

func (a *AxlResponse) responseReturnData() error {
	log.WithField("id", a.id).Debugf("AXL response status [%s]", a.response.Status)
	bodies, err := ioutil.ReadAll(a.response.Body)
	_ = a.response.Body.Close()
	a.body = ""

	if err != nil {
		log.WithField("id", a.id).Errorf("problem get AXL body from response - [%s]", err)
		return err
	}
	if a.statusCode > 299 {
		a.getFailurePart(string(bodies))
	} else {
		a.getReturnPart(string(bodies))
	}
	log.WithField("id", a.id).Trace("Body read success")
	return nil
}

func (a *AxlResponse) getReturnPart(body string) {
	a.getBetween(body, "<return>", "</return>", "<return/>")
}

func (a *AxlResponse) getFailurePart(body string) {
	a.getBetween(body, "<soapenv:Fault>", "</soapenv:Fault>", "<soapenv:Fault/>")
}

func (a *AxlResponse) getBetween(body string, start string, end string, short string) {
	if strings.Contains(body, start) {
		body = body[strings.Index(body, start):]
	} else if strings.Index(body, short) > -1 {
		a.body = short
		return
	}
	a.body = ""
	if !strings.Contains(body, end) {
		return
	}
	a.body = body[:strings.Index(body, end)] + end
}

func (a *AxlResponse) GetResponseBody() string {
	if a.response == nil {
		return a.body
	}
	err := a.responseReturnData()
	if err != nil {
		a.err = err
	}
	return a.body
}
