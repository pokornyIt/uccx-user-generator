package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type AxlRequest struct {
	id       string
	sequence int
	server   *AxlServer
	request  *http.Request
}

func (a *AxlRequest) getCmVersionBody() string {
	return fmt.Sprintf(AxlXmlHeaderFormat+AxlCcmVersion, a.server.dbVersion, a.sequence)
}

func (a *AxlRequest) getSqlRequestBody(sql string) string {
	return fmt.Sprintf(AxlXmlHeaderFormat+AxlSqlRequest, a.server.dbVersion, a.sequence, sql)
}

func (a *AxlRequest) DbVersionRequest() *AxlResponse {
	sql := a.getCmVersionBody()
	return a.doAxlRequest(sql)
}

func (a *AxlRequest) setHeader() {
	a.request.Header.Set("Content-Type", "text/xml")
	a.request.Header.Set("User-Agent", appVersion())
	a.request.Header.Set("Accept", "*/*")
	a.request.Header.Set("Cache-Control", "no-cache")
	a.request.Header.Set("Pragma", "no-cache")
	a.request.Host = a.server.server
	a.request.SetBasicAuth(a.server.user, a.server.pwd)
}

func (a *AxlRequest) doAxlRequest(body string) *AxlResponse {
	log.WithField("id", a.id).Tracef("process AXL request to [%s]", a.server.server)

	req, err := http.NewRequest("POST", a.server.getUrl(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.WithField("id", a.id).Errorf("problem create new POST request to AXL [%s] - [%s", a.server.server, err)
		return a.NewAxlResponse(nil, err, fmt.Sprintf("problem create new POST request to AXL [%s]", a.server.server))
	}
	a.request = req
	log.WithField("id", a.id).Trace("success create new AXL request")
	return a.finishRequest()
}

func (a *AxlRequest) finishRequest() *AxlResponse {
	if !a.server.isAuthValid {
		return a.NewAxlResponse(nil, fmt.Errorf("user not authorize to access AXL"), "user not authorize to access AXL")
	}
	a.setHeader()
	resp, err := a.server.httpClient.Do(a.request)
	if err != nil {
		log.WithField("id", a.id).Errorf("problem process request  to server [%s] - [%s].", a.server.server, err)
		return a.NewAxlResponse(nil, err, "problem "+a.request.Method+" response")
	}
	return a.NewAxlResponse(resp, nil, "")
}

func (a *AxlRequest) NewAxlResponse(r *http.Response, e error, message string) *AxlResponse {
	c := new(AxlResponse)
	c.id = a.id
	c.response = r
	c.err = e
	c.lastMessage = message
	if r != nil {
		c.statusCode = r.StatusCode
		c.statusMessage = r.Status
	} else {
		c.statusCode = 500
		c.statusMessage = "500 Problem Connect to server"
	}
	log.WithField("id", c.id).Debugf("Create new response for request to [%s]", a.server.getUrl())
	return c
}
