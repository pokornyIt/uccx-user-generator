package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

const (
	AxlSqlUser = `
SELECT eu.pkid, eu.userid, eu.status, eunp.uccx FROM enduser eu
	LEFT OUTER JOIN (SELECT fkenduser, max(CASE tkdnusage WHEN 2 THEN tkdnusage ELSE null END) is not null AS uccx
    	FROM endusernumplanmap GROUP BY fkenduser) AS eunp ON eunp.fkenduser = eu.pkid
WHERE tkuserprofile=1
`
)

type axlEndUserList struct {
	XMLName xml.Name     `xml:"return"`
	Rows    []axlEndUser `xml:"row"`
}

type axlEndUser struct {
	XMLName   xml.Name `xml:"row"`
	ClusterId string   `xml:"pkid"`
	UserId    string   `xml:"userid"`
	Status    int      `xml:"status"`
	Uccx      string   `xml:"uccx"`
}

var (
	axlEndUsersList  *axlEndUserList
	axlEndUsersMutex sync.Mutex
)

func asyncAxlReadUsers(server *AxlServer, wd *sync.WaitGroup) {
	req := server.newSoapRequest()
	body := req.getSqlRequestBody(AxlSqlUser)
	resp := req.doAxlRequest(body)
	fmt.Printf(".... finish AXL user request\r\n")
	if resp.err != nil {
		log.Errorf("problem read data form AXL %s", resp.err)
	} else {
		body = resp.GetResponseBody()
		var data axlEndUserList
		err := xml.Unmarshal([]byte(body), &data)
		if err != nil {
			log.Errorf("problme unmarshal AXl response for get user - %s", err)
		} else {
			log.Infof("from AXl read %d users", len(data.Rows))
			axlEndUsersMutex.Lock()
			axlEndUsersList = &data
			axlEndUsersMutex.Unlock()
		}
	}
	wd.Done()
}
