package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type AxlServer struct {
	server      string
	user        string
	pwd         string
	dbVersion   string
	sequence    int
	timeout     int
	isAuthValid bool
	httpClient  *http.Client
	mutex       sync.Mutex
}

func newAxlServer() *AxlServer {
	c := &AxlServer{
		server:      *Config.axlServer,
		user:        *Config.axlUserName,
		pwd:         *Config.axlUserPassword,
		timeout:     *Config.timeOut,
		sequence:    1,
		dbVersion:   "",
		isAuthValid: true,
		httpClient:  nil,
		mutex:       sync.Mutex{},
	}
	c.getClient()
	_, err := c.GetDbVersion()
	if err != nil {
		log.Fatalf("program connect to not supported CUCM server")
		programExit(1)
	}
	return c
}

func (s *AxlServer) getClient() *http.Client {
	if s.httpClient == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		s.httpClient = &http.Client{Timeout: time.Duration(s.timeout) * time.Second, Transport: tr}
		log.Tracef("prepared HTTP client for server [%s]", s.server)
	}
	return s.httpClient
}

func (s *AxlServer) getUrl() string {
	return fmt.Sprintf(AxlUrlFormat, s.server)
}

func (s *AxlServer) newSoapRequest() *AxlRequest {
	r := AxlRequest{
		id:       AxlIdPrefix + RandomString(),
		server:   s,
		request:  nil,
		sequence: s.getNewSequenceId(),
	}
	log.WithField("id", r.id).Tracef("create new request with id [%s]", r.id)
	return &r
}

func (s *AxlServer) getNewSequenceId() int {
	s.mutex.Lock()
	s.sequence = s.sequence + 1
	s.mutex.Unlock()
	return s.sequence
}

func (s *AxlServer) GetDbVersion() (string, error) {
	log.Tracef("try get CUCM Db version for server [%s]", s.server)
	var err error
	if s.dbVersion == "" {
		for i := 0; i < len(AxlDbVersions); i++ {
			s.dbVersion = AxlDbVersions[i]
			log.Debugf("test CUCM version [%s]", s.dbVersion)
			request := s.newSoapRequest()
			resp := request.DbVersionRequest()
			if resp.err != nil {
				s.dbVersion = AxlDbVersionError
				resp.Close()
				return resp.lastMessage, resp.err
			}
			if resp.statusCode == 599 {
				s.dbVersion = AxlDbVersionError
				resp.Close()
				continue
			}
			if resp.statusCode == 401 {
				s.isAuthValid = false
				resp.Close()
				log.Errorf("problem with AXL authorization HTTP status [401]")
				return "Problem with AXL authorization", fmt.Errorf(resp.statusMessage)
			}
			if resp.statusCode == 200 {
				v, err := VersionData(resp.GetResponseBody())
				if err == nil {
					log.Infof("actual AXL version [%s], DbVersion [%s]", v.Version, v.GetDbVersion())
					s.dbVersion = v.GetDbVersion()
					resp.Close()
					break
				}
				log.Warningf("problem convert XML data to version structure")
			}
			s.dbVersion = AxlDbVersionError
			resp.Close()
		}
	}
	return s.dbVersion, err
}
