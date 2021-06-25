package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type CcxResponse struct {
	id            string
	url           string
	response      *http.Response
	err           error
	lastMessage   string
	body          string
	statusCode    int
	statusMessage string
}

func (r *CcxResponse) close() {
	if r.response != nil && r.response.Body != nil {
		_ = r.response.Body.Close()
	}
	r.response = nil
}

func (r *CcxResponse) responseReturnData() error {
	log.WithField("id", r.id).Tracef("response status is [%s]", r.response.Status)
	bodies, err := ioutil.ReadAll(r.response.Body)
	_ = r.response.Body.Close()
	r.body = string(bodies)

	if err != nil {
		log.WithField("id", r.id).Errorf("problem get body from response with message [%s].", err)
		return err
	}
	log.WithField("id", r.id).Tracef("success read [%d] chars of body", len(r.body))
	return nil
}

func (r *CcxResponse) GetResponseBody() string {
	if r.response == nil {
		return r.body
	}
	err := r.responseReturnData()
	if err != nil {
		r.err = err
	}
	return r.body
}
