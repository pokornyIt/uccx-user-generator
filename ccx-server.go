package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type CcxServer struct {
	server     string
	user       string
	pwd        string
	timeout    int
	httpClient *http.Client
}

func NewCcxServer() *CcxServer {
	c := &CcxServer{
		server:     *Config.ccServer,
		user:       *Config.ccUserName,
		pwd:        *Config.ccPassword,
		timeout:    *Config.timeOut,
		httpClient: nil,
	}
	c.getClient()
	return c
}

func (c *CcxServer) getClient() *http.Client {
	if c.httpClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.httpClient = &http.Client{Timeout: time.Duration(c.timeout) * time.Second, Transport: tr}
		log.Tracef("prepared HTTP client for server [%s]", c.server)
	}
	return c.httpClient
}

func (c *CcxServer) getUrl(path string) string {
	if strings.HasPrefix(path, fmt.Sprintf("https://%s", c.server)) {
		return path
	}
	if strings.HasPrefix(path, CcxUrlMainPart) {
		return fmt.Sprintf("https://%s%s", c.server, path)
	}

	return fmt.Sprintf("https://%s%s%s", c.server, CcxUrlMainPart, path)
}

func (c *CcxServer) newRestRequest(url string) *CcxRequest {
	r := CcxRequest{
		id:      CcxIdPrefix + RandomString(),
		server:  c,
		url:     url,
		request: nil,
	}
	log.WithField("id", r.id).Tracef("create new request with id [%s]", r.id)
	return &r
}
