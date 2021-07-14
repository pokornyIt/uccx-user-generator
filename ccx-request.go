package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type CcxRequest struct {
	id      string
	server  *ccxServer
	url     string
	request *http.Request
}

type ccxRefObject struct {
	Name   string `json:"@name"`
	RefUrl string `json:"refURL"`
}

func (s *CcxRequest) setHeader(isPost bool) {
	if isPost {
		s.request.Header.Set("Content-Type", "application/json")
	}
	s.request.Header.Set("User-Agent", appVersion())
	s.request.Header.Set("Accept", "application/json")
	s.request.Header.Set("Cache-Control", "no-cache")
	s.request.Header.Set("Pragma", "no-cache")
	s.request.Header.Set("X-User-Generator-Id", s.id)
	s.request.Host = s.server.server
	s.request.SetBasicAuth(s.server.user, s.server.pwd)
}

func (s *CcxRequest) doGetRequest() *CcxResponse {
	log.WithField("id", s.id).Tracef("process GET request to [%s]", s.url)
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		log.WithField("id", s.id).Errorf("problem create GET request to [%s] with error %s", s.url, err)
		return s.newRestResponse(nil, err, fmt.Sprintf("problem create request to [%s] with error %s", s.url, err))
	}
	s.request = req
	s.setHeader(false)
	return s.finishRequest()
}

func (s CcxRequest) doPostRequest(body string) *CcxResponse {
	return s.doDataSendRequest(body, "POST")
}

func (s *CcxRequest) doDeleteRequest(body string) *CcxResponse {
	return s.doDataSendRequest(body, "DELETE")
}

func (s *CcxRequest) doDataSendRequest(body string, operation string) *CcxResponse {
	log.WithField("id", s.id).Tracef("process %s request to [%s]", operation, s.url)
	req, err := http.NewRequest(operation, s.url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.WithField("id", s.id).Errorf("problem create %s request to [%s] with error %s", operation, s.url, err)
		return s.newRestResponse(nil, err, fmt.Sprintf("problem create request to [%s] with error %s", s.url, err))
	}
	s.request = req
	s.setHeader(true)
	return s.finishRequest()
}

func (s *CcxRequest) finishRequest() *CcxResponse {
	resp, err := s.server.httpClient.Do(s.request)
	if err != nil {
		log.WithField("id", s.id).Errorf("problem process request to [%s] with error %s", s.url, err)
		programExit(1)
		//return s.newRestResponse(nil, err, fmt.Sprintf("problem process request to [%s] with error %s", s.url, err))
	}
	return s.newRestResponse(resp, nil, "success create request")
}

func (s *CcxRequest) newRestResponse(response *http.Response, err error, message string) *CcxResponse {
	c := new(CcxResponse)
	c.id = s.id
	c.url = s.url
	c.response = response
	c.err = err
	c.lastMessage = message
	if response != nil {
		c.statusCode = response.StatusCode
		c.statusMessage = response.Status
	} else {
		c.statusCode = http.StatusInternalServerError
		c.statusMessage = "500 Problem Connect to server"
	}
	log.WithField("id", s.id).Debugf("create new response for request to [%s]", s.url)
	return c
}
