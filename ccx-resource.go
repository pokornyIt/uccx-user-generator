package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sync"
)

type ccxResourceList struct {
	Type     string        `json:"@type"`
	Resource []ccxResource `json:"resource"`
}

type ccxResource struct {
	Self                  string          `json:"self"`
	UserId                string          `json:"userID"`
	FirstName             string          `json:"firstName"`
	LastName              string          `json:"lastName"`
	Extension             string          `json:"extension"`
	Alias                 string          `json:"alias"`
	ResourceGroup         ccxRefObject    `json:"resourceGroup"`
	SkillMap              ccxSkillMap     `json:"skillMap"`
	AutoAvailable         bool            `json:"autoAvailable"`
	Type                  int             `json:"type"`
	Team                  ccxRefObject    `json:"team"`
	PrimarySupervisorOf   ccxSupervisorOf `json:"primarySupervisorOf"`
	SecondarySupervisorOf ccxSupervisorOf `json:"secondarySupervisorOf"`
}

type ccxSkillMap struct {
	SkillCompetency []ccxSkillCompetency `json:"skillCompetency"`
}

type ccxSkillCompetency struct {
	CompetenceLevel  int          `json:"competencelevel"`
	SkillNameUriPair ccxRefObject `json:"skillNameUriPair"`
}

type ccxSupervisorOf struct {
	SupervisorOfTeamName []ccxRefObject `json:"supervisorOfTeamName"`
}

var (
	ccxUserActiveList  *ccxResourceList
	ccxUserActiveMutex sync.Mutex
)

func asyncCcxResourceList(server *ccxServer, wg *sync.WaitGroup) {
	user := &ccxResourceList{Resource: nil}

	url := server.getUrl(CcxResourcePath)
	request := server.newRestRequest(url)
	response := request.doGetRequest()
	fmt.Printf(".... finish CCX user request\r\n")
	if response.err != nil {
		log.Error(response.err)
	} else {
		err := json.Unmarshal([]byte(response.GetResponseBody()), &user)
		if err != nil {
			log.Errorf("problem when convert CCX resource request - %s", err)
			user = &ccxResourceList{Resource: nil}
		}
	}

	if user.hasResources() {
		ccxGen := user.getGeneratedUsers()
		log.Infof("from CCX server read %d users and %d is generated", len(user.Resource), len(ccxGen))
	} else {
		log.Errorf("problem communicate with CCX server")
	}
	ccxUserActiveMutex.Lock()
	ccxUserActiveList = user
	ccxUserActiveMutex.Unlock()
	wg.Done()
}

func (c *ccxResourceList) getGeneratedUsers() []ccxResource {
	var data []ccxResource
	if !c.hasResources() {
		return data
	}
	var re = regexp.MustCompile(CcxUserNameRegex)
	for i := 0; i < len(c.Resource); i++ {
		if re.MatchString(c.Resource[i].UserId) {
			data = append(data, c.Resource[i])
		}
	}
	return data
}

func (c *ccxResourceList) hasResources() bool {
	return c.Resource != nil && len(c.Resource) > 0
}
