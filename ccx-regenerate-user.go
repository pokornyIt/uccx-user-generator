package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

type ccxForceResponse struct {
	Response  string             `json:"Response"`
	AgentInfo []ccxResourceForce `json:"Agent_Info"`
}

type ccxResourceForce struct {
	UserID              string   `json:"User_ID"`
	ResourceFullName    string   `json:"Resource_FullName"`
	ResourceGroupName   string   `json:"ResourceGroup_Name"`
	IPCCExtension       string   `json:"IPCC_Extension"`
	TeamID              int      `json:"Team_ID"`
	TeamName            string   `json:"Team_Name"`
	AgentState          int      `json:"Agent_State"`
	AutoAvailable       bool     `json:"Auto_Available"`
	NumericID           int      `json:"Numeric_ID"`
	Order               int      `json:"Order"`
	ResourceGroupID     int      `json:"ResourceGroup_ID"`
	FirstName           string   `json:"First_Name"`
	Type                int      `json:"Type"`
	LastName            string   `json:"Last_Name"`
	PrimarySupervisor   []int    `json:"PrimarySupervisor"`
	SecondarySupervisor []int    `json:"SecondarySupervisor"`
	AssignedSkillList   []string `json:"Assigned_SkillList"`
}

func CcxResourceForceSync() (*ccxForceResponse, error) {
	timeStart := time.Now()
	server := newCcxForceServer()
	var user ccxForceResponse

	url := server.getUrlForce()
	request := server.newRestRequest(url)
	response := request.doGetRequest()
	timeDuration := time.Since(timeStart)
	log.WithField("id", request.id).Infof("CCX force user update run [%s]", timeDuration.String())
	if response.err != nil {
		log.Error(response.err)
		return nil, response.err
	}
	body, err := response.GetResponseBody()
	if err != nil {
		log.Error(response.err)
		return nil, response.err
	}
	err = json.Unmarshal([]byte(body), &user)
	if err != nil {
		log.WithField("id", request.id).Errorf("problem when convert CCX resource request - %s", err)
		response.storeResponse()
		return nil, err
	}
	return &user, nil
}

func (r *ccxResourceForce) isGenerated() bool {
	return axlUserRegex.MatchString(r.UserID)
}

func (f *ccxForceResponse) hasResources() bool {
	return f.AgentInfo != nil && len(f.AgentInfo) > 0
}

func (f *ccxForceResponse) getGenerated() []ccxResourceForce {
	var data []ccxResourceForce
	if !f.hasResources() {
		return data
	}
	for i := 0; i < len(f.AgentInfo); i++ {
		if f.AgentInfo[i].isGenerated() {
			data = append(data, f.AgentInfo[i])
		}
	}
	return data
}
