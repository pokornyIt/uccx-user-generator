package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sync"
)

type ccxTeamList struct {
	Type string    `json:"@type"`
	Team []ccxTeam `json:"team"`
}

type ccxTeam struct {
	Self                 string                  `json:"self"`
	TeamId               int                     `json:"teamId"`
	TeamName             string                  `json:"teamname"`
	PrimarySupervisor    ccxRefObject            `json:"primarySupervisor"`
	SecondarySupervisors ccxSecondarySupervisors `json:"secondarySupervisors"`
}

type ccxSecondarySupervisors struct {
	SecondarySupervisor []ccxRefObject `json:"secondrySupervisor"`
}

var (
	ccxTeamActiveList *ccxTeamList
	ccxTeamListMutex  sync.Mutex
)

func asyncCcxTeamList(server *ccxServer, wg *sync.WaitGroup) {
	team := &ccxTeamList{Team: nil}

	url := server.getUrl(CcxTeamPath)
	request := server.newRestRequest(url)
	response := request.doGetRequest()
	fmt.Printf(".... finish CCX team request\r\n")
	if response.err != nil {
		log.Error(response.err)
	} else {
		err := json.Unmarshal([]byte(response.GetResponseBody()), &team)
		if err != nil {
			log.Errorf("problem when convert CCX teams request - %s", err)
			team = &ccxTeamList{Team: nil}
		}
	}

	if team.hasTeams() {
		ccxTeamGen := team.getGeneratedTeams()
		log.Infof("from CCX server read %d teams and %d is generated", len(team.Team), len(ccxTeamGen))
	} else {
		log.Errorf("problem collect data from CCX server")
	}
	ccxTeamListMutex.Lock()
	ccxTeamActiveList = team
	ccxTeamListMutex.Unlock()
	wg.Done()
}

func (t *ccxTeamList) getGeneratedTeams() []ccxTeam {
	var data []ccxTeam
	if !t.hasTeams() {
		return data
	}
	var re = regexp.MustCompile(CcxUserNameRegex)
	for i := 0; i < len(t.Team); i++ {
		if re.MatchString(t.Team[i].TeamName) {
			data = append(data, t.Team[i])
		}
	}
	return data
}

func (t *ccxTeamList) hasTeams() bool {
	return t.Team != nil && len(t.Team) > 0
}
