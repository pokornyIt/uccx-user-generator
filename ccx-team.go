package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"sync"
)

type ccxTeamList struct {
	Type string    `json:"@type"`
	Team []ccxTeam `json:"team"`
}

type ccxTeam struct {
	Self                 string                  `json:"self"`
	TeamId               int                     `json:"teamId,omitempty"`
	TeamName             string                  `json:"teamname"`
	PrimarySupervisor    ccxRefObject            `json:"primarySupervisor,omitempty"`
	SecondarySupervisors ccxSecondarySupervisors `json:"secondarySupervisors,omitempty"`
}

type ccxTeamDetail struct {
	Self              string          `json:"self"`
	TeamId            int             `json:"teamId,omitempty"`
	TeamName          string          `json:"teamname"`
	PrimarySupervisor ccxRefObject    `json:"primarySupervisor,omitempty"`
	Resources         ccxTeamResource `json:"resources,omitempty"`
}

type ccxTeamResource struct {
	Resource []ccxRefObject `json:"resource,omitempty"`
}

type ccxSecondarySupervisors struct {
	SecondarySupervisor []ccxRefObject `json:"secondrySupervisor,omitempty"`
}

func asyncCcxTeamList(server *ccxServer, wg *sync.WaitGroup) {
	team := &ccxTeamList{Team: nil}

	url := server.getUrl(CcxTeamPath)
	request := server.newRestRequest(url)
	response := request.doGetRequest()
	if response.err != nil {
		log.Error(response.err)
	} else {
		body, err := response.GetResponseBody()
		if err != nil {
			log.Error(response.err)
		} else {
			err = json.Unmarshal([]byte(body), &team)
			if err != nil {
				log.WithField("id", request.id).Errorf("problem when convert CCX teams request - %s", err)
				response.storeResponse()
				team = &ccxTeamList{Team: nil}
			}
		}
	}

	if team.hasTeams() {
		ccxTeamGen := team.getGeneratedTeams()
		log.WithField("id", request.id).Infof("from CCX server read %d teams and %d is generated", len(team.Team), len(ccxTeamGen))
	} else {
		log.WithField("id", request.id).Errorf("problem collect data from CCX server")
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
	for i := 0; i < len(t.Team); i++ {
		if t.Team[i].isGenerated() {
			data = append(data, t.Team[i])
		}
	}
	return data
}

func (t *ccxTeamList) hasTeams() bool {
	return t.Team != nil && len(t.Team) > 0
}

func (t *ccxTeamList) getTeamToDelete(no int) []ccxTeam {
	sortTeam := t.getGeneratedTeams()
	if len(sortTeam) < 1 {
		return nil
	}
	sort.Slice(sortTeam[:], func(i, j int) bool {
		return sortTeam[i].TeamName > sortTeam[j].TeamName
	})
	var deleteTeams []ccxTeam
	for i := 0; i < no && i < len(sortTeam); i++ {
		deleteTeams = append(deleteTeams, sortTeam[i])
	}
	return deleteTeams
}

func (t *ccxTeamList) removeTeamName(name string) {
	for i := 0; i < len(t.Team); i++ {
		if t.Team[i].TeamName == name {
			t.Team = append(t.Team[:i], t.Team[i+1:]...)
			return
		}
	}
}

func (t *ccxTeamList) getCcxTeamByName(name string) *ccxTeam {
	for i := 0; i < len(t.Team); i++ {
		if t.Team[i].TeamName == name {
			return &t.Team[i]
		}
	}
	return nil
}

func (t *ccxTeam) isGenerated() bool {
	return ccxTeamRegex.MatchString(t.TeamName)
}

func (t *ccxTeam) teamUrl(server *ccxServer) string {
	return fmt.Sprintf("%s/%d", server.getUrl(CcxTeamPath), t.TeamId)
}

func (t *ccxTeam) deleteTeam(server *ccxServer) error {
	if !t.isGenerated() {
		log.Errorf("delete request for team [%s] not sucess, team not generated", t.TeamName)
		return fmt.Errorf("delete request for team [%s] not sucess, team not generated", t.TeamName)
	}
	url := t.teamUrl(server)
	request := server.newRestRequest(url)
	response := request.doDeleteRequest("")
	if response.err != nil {
		log.Error(response.err)
		return response.err
	}
	if response.statusCode != http.StatusOK {
		log.Errorf("delete request for team [%s] not sucess, status code is [%d]", t.TeamName, response.statusCode)
		return fmt.Errorf("delete request for team [%s] not sucess, status code is [%d]", t.TeamName, response.statusCode)
	}
	log.Debugf("delete request for team [%s] sucessed", t.TeamName)
	return nil
}

func (t *ccxTeamList) teamNameExist(name string) bool {
	for _, team := range t.Team {
		if team.TeamName == name {
			return true
		}
	}
	return false
}

func (t *ccxTeamList) createUpdateTeam(server *ccxServer, teamId int, resource []ccxRefObject) error {
	var supervisor ccxRefObject
	supSet := false
	for i := 0; i < len(t.Team); i++ {
		if len(t.Team[i].PrimarySupervisor.RefUrl) > 0 {
			supervisor = t.Team[i].PrimarySupervisor
			supSet = true
			break
		}
	}
	if !supSet {
		log.Errorf("problem get supervisor from team list for create team ID [%s]", fmt.Sprintf(CcxTeamNameFormat, teamId))
		return fmt.Errorf("problem get supervisor from team list for create team ID [%s]", fmt.Sprintf(CcxTeamNameFormat, teamId))
	}
	newTeam := t.newCcxTeam(server, teamId, supervisor, resource)
	data, err := json.Marshal(newTeam)
	if err != nil {
		log.Errorf("problem create new team [%s] with error [%s]", newTeam.TeamName, err)
		return err
	}
	dataString := string(data)
	var response *CcxResponse
	if t.teamNameExist(newTeam.TeamName) {
		url := fmt.Sprintf("%s/%d", server.getUrl(CcxTeamPath), newTeam.TeamId)
		request := server.newRestRequest(url)
		response = request.doPutRequest(dataString)
	} else {
		url := server.getUrl(CcxTeamPath)
		request := server.newRestRequest(url)
		response = request.doPostRequest(dataString)
	}
	if response.err != nil {
		log.Errorf("problem create team [%s] with status [%d] and error [%s]", newTeam.TeamName, response.statusCode, response.err)
		return fmt.Errorf("problem create team [%s] with status [%d] and error [%s]", newTeam.TeamName, response.statusCode, response.err)
	}
	if response.statusCode != http.StatusCreated && response.statusCode != http.StatusOK {
		log.Errorf("problem create team %s with status %d and message %s", newTeam.TeamName, response.statusCode, response.statusMessage)
		return fmt.Errorf("problem create team %s with status %d and message %s", newTeam.TeamName, response.statusCode, response.statusMessage)
	}
	return nil
}

func (t *ccxTeamList) newCcxTeam(server *ccxServer, teamId int, supervisor ccxRefObject, resource []ccxRefObject) *ccxTeamDetail {
	var c ccxTeamDetail
	teamName := fmt.Sprintf(CcxTeamNameFormat, teamId)
	oldTeam := t.getCcxTeamByName(teamName)
	if oldTeam != nil {
		c = ccxTeamDetail{
			Self:              oldTeam.Self,
			TeamId:            oldTeam.TeamId,
			TeamName:          teamName,
			PrimarySupervisor: supervisor,
			Resources:         ccxTeamResource{resource},
		}
	} else {
		c = ccxTeamDetail{
			Self:              server.getUrl(CcxTeamPath),
			TeamId:            1,
			TeamName:          teamName,
			PrimarySupervisor: supervisor,
			Resources:         ccxTeamResource{resource},
		}
	}
	return &c
}
