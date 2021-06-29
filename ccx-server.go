package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type ccxServer struct {
	server     string
	user       string
	pwd        string
	timeout    int
	httpClient *http.Client
}

func newCcxServer() *ccxServer {
	c := &ccxServer{
		server:     *Config.ccServer,
		user:       *Config.ccUserName,
		pwd:        *Config.ccPassword,
		timeout:    *Config.timeOut,
		httpClient: nil,
	}
	c.getClient()
	return c
}

func (c *ccxServer) getClient() *http.Client {
	if c.httpClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.httpClient = &http.Client{Timeout: time.Duration(c.timeout) * time.Second, Transport: tr}
		log.Tracef("prepared HTTP client for server [%s]", c.server)
	}
	return c.httpClient
}

func (c *ccxServer) getUrl(path string) string {
	if strings.HasPrefix(path, fmt.Sprintf("https://%s", c.server)) {
		return path
	}
	if strings.HasPrefix(path, CcxUrlMainPart) {
		return fmt.Sprintf("https://%s%s", c.server, path)
	}

	return fmt.Sprintf("https://%s%s%s", c.server, CcxUrlMainPart, path)
}

func (c *ccxServer) newRestRequest(url string) *CcxRequest {
	r := CcxRequest{
		id:      CcxIdPrefix + RandomString(),
		server:  c,
		url:     url,
		request: nil,
	}
	log.WithField("id", r.id).Tracef("create new request with id [%s] for [%s]", r.id, url)
	return &r
}
