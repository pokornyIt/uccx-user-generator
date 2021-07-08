package main

import (
	"encoding/xml"
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
	return data
}

func (l *axlEndUserList) hasUsers() bool {
	return l.Rows != nil && len(l.Rows) > 0
}

func (l *axlEndUserList) disableUser(user axlEndUser) error {
	log.Tracef("disable user [%s]", user.UserId)
	if user.Status == 0 {
		log.Debugf("user [%s] is disabled", user.UserId)
		return nil
	}
	return nil
}

func (l *axlEndUserList) enableUser(user axlEndUser) error {
	log.Tracef("enable user [%s]", user.UserId)
	if user.Status == 1 {
		log.Debugf("user [%s] is enabled", user.UserId)
		return nil
	}
	return nil
}

func (a *axlEndUser) isGenerated() bool {
	return axlUserRegex.MatchString(a.UserId)
}

func (a *axlEndUser) isEnableForCcx() bool {
	return a.Uccx != ""
}
