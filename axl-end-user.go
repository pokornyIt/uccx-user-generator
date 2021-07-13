package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

const (
	AxlSqlUser = `
SELECT eu.pkid, eu.userid, eu.status, eunp.uccx, eudm.fkdevice, dnp.fknumplan, np.dnorpattern 
FROM enduser AS eu
	LEFT OUTER JOIN 
    (SELECT fkenduser, max(CASE tkdnusage WHEN 2 THEN tkdnusage ELSE null END) is not null AS uccx FROM endusernumplanmap GROUP BY fkenduser) AS eunp ON eunp.fkenduser = eu.pkid
    LEFT OUTER JOIN 
    (SELECT fkenduser, Min(fkdevice) fkdevice FROM enduserdevicemap GROUP BY fkenduser) AS eudm ON eudm.fkenduser =  eu.pkid
    LEFT OUTER JOIN
    (SELECT DISTINCT fkdevice, fknumplan FROM devicenumplanmap WHERE fkdevice is not null and fknumplan is not null AND numplanindex=1) AS dnp ON dnp.fkdevice = eudm.fkdevice
    LEFT OUTER JOIN
    (SELECT pkid, dnorpattern FROM numplan WHERE dnorpattern is not null) AS np ON np.pkid = dnp.fknumplan
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
	Ccx       string   `xml:"uccx"`
	DeviceId  string   `xml:"fkdevice"`
	NumberId  string   `xml:"fknumplan"`
	LineDn    string   `xml:"dnorpattern"`
}

type axlCcxSwitchResponse struct {
	XMLName xml.Name `xml:"return"`
	Id      string   `xml:",chardata"`
}

func asyncAxlReadUsers(server *AxlServer, wd *sync.WaitGroup) {
	req := server.newSoapRequest()
	body := req.getSqlRequestBody(AxlSqlUser)
	resp := req.doAxlRequest(body)
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

func (l *axlEndUserList) getGeneratedUsers() []axlEndUser {
	var data []axlEndUser
	if !l.hasUsers() {
		return data
	}
	for i := 0; i < len(l.Rows); i++ {
		if l.Rows[i].isGenerated() {
			data = append(data, l.Rows[i])
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].UserId < data[j].UserId
	})
	return data
}

func (l *axlEndUserList) getGeneratedCcxUsers() []axlEndUser {
	var data []axlEndUser
	if !l.hasUsers() {
		return data
	}
	for i := 0; i < len(l.Rows); i++ {
		if l.Rows[i].isGenerated() && l.Rows[i].isEnableForCcx() {
			data = append(data, l.Rows[i])
		}
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].UserId < data[j].UserId
	})
	return data
}

func (l *axlEndUserList) getNeedEnabledUsers() int {
	allEnabled := 0
	for i := 0; i < len(l.Rows); i++ {
		if !l.Rows[i].isGenerated() && l.Rows[i].isEnableForCcx() {
			allEnabled++
		}
	}
	if allEnabled >= *Config.finalNumber {
		return 0
	}
	return *Config.finalNumber - allEnabled
}

func (l *axlEndUserList) hasUsers() bool {
	return l.Rows != nil && len(l.Rows) > 0
}

func (a *axlEndUser) isGenerated() bool {
	return axlUserRegex.MatchString(a.UserId)
}

func (a *axlEndUser) isEnableForCcx() bool {
	return a.Ccx != ""
}

func (a *axlEndUser) switchCcxLine(server *AxlServer, setLine bool, wg *sync.WaitGroup) {
	log.WithField("user", a.UserId).Infof("try change to %t, for user %s", setLine, a.UserId)
	req := server.newSoapRequest()
	var body string
	if a.isEnableForCcx() == setLine {
		log.Tracef("AXL user %s is in required state", a.UserId)
		wg.Done()
		return
	}
	if setLine {
		body = req.getUserSetCcxRequestBody(a.ClusterId, a.NumberId, a.LineDn)
	} else {
		body = req.getUserRemoveCcxRequestBody(a.ClusterId)
	}
	resp := new(AxlResponse)
	resp.err = fmt.Errorf("request not initisalited")
	for i := 0; i < AxlRepeatRequestCount; i++ {
		resp = req.doAxlRequest(body)
		if resp.err != nil {
			break
		}
		if resp.statusCode == 200 {
			break
		}
		time.Sleep(AxlRepeatWaitSeconds * time.Second)
	}
	if resp.err != nil {
		log.WithField("user", a.UserId).Errorf("problem read data form AXL %s", resp.err)
	} else {
		body = resp.GetResponseBody()
		var data axlCcxSwitchResponse
		err := xml.Unmarshal([]byte(body), &data)
		if err != nil {
			log.WithFields(log.Fields{"user": a.UserId, "body": body, "code": resp.statusCode}).Errorf("problme unmarshal AXl response for update CCX line - %s", err)
		} else {
			if setLine {
				log.Debugf("AXL set CCX from user %s", a.UserId)
				a.Ccx = "t"
			} else {
				log.Debugf("AXL remove CCX from user %s", a.UserId)
				a.Ccx = ""
			}
		}
	}
	wg.Done()
}
