package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"
)

var (
	Version      string
	Revision     string
	Branch       string
	BuildUser    string
	BuildDate    string
	ccxApiServer *ccxServer
	axlServer    *AxlServer
	roundTime    statData

	GoVersion = runtime.Version() // build GO version
	infoTmpl  = `
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
  build user:       {{.buildUser}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
  platform:         {{.platform}}
`
	src           = rand.NewSource(time.Now().UnixNano()) // randomize base string
	maxRandomSize = 10                                    // required size of random string
)

func VersionDetail() string {
	m := map[string]string{
		"program":   ApplicationName,
		"version":   Version,
		"revision":  Revision,
		"branch":    Branch,
		"buildUser": BuildUser,
		"buildDate": BuildDate,
		"goVersion": GoVersion,
		"platform":  runtime.GOOS + "/" + runtime.GOARCH,
	}
	t := template.Must(template.New("version").Parse(infoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

func readActualData() {
	var wg sync.WaitGroup

	wg.Add(1)
	go asyncCcxResourceList(ccxApiServer, &wg)

	wg.Add(1)
	go asyncAxlReadUsers(axlServer, &wg)

	wg.Add(1)
	go asyncCcxTeamList(ccxApiServer, &wg)

	wg.Wait()
	log.Infof("Finish initial data read from all sources")

	log.Infof("finish read %d/%d CCX teams", len(ccxTeamActiveList.Team), len(ccxTeamActiveList.getGeneratedTeams()))
	log.Infof("finish read %d/%d CCX users", len(ccxUserActiveList.Resource), len(ccxUserActiveList.getGeneratedUsers()))
	log.Infof("finish read %d/%d CUCM users %d enabled for CCX", len(axlEndUsersList.Rows), len(axlEndUsersList.getGeneratedUsers()), len(axlEndUsersList.getGeneratedCcxUsers()))
}

type chanData struct {
	Id      int
	SetLine bool
}

func setAgents() {
	data := axlEndUsersList.getGeneratedUsers()
	nowAxlEnabled := len(axlEndUsersList.getGeneratedCcxUsers())
	nowCcxEnabled := len(ccxUserActiveList.Resource)
	if nowAxlEnabled > nowCcxEnabled {
		log.Error("problem CCX not synchronized with AXL")
		programExit(1)
	}
	operation := *Config.finalNumber > nowCcxEnabled
	var needUpdate int
	if *Config.finalNumber == 0 {
		needUpdate = nowAxlEnabled
	} else {
		if operation { // add
			needUpdate = *Config.finalNumber - nowCcxEnabled
		} else { //remove
			needUpdate = nowCcxEnabled - *Config.finalNumber
			if needUpdate > nowAxlEnabled { // not generated more than need remove
				needUpdate = nowAxlEnabled
			}
		}
	}
	roundTime.Direction = operation
	roundTime.ExpectUpdates = needUpdate

	setCcx := 0
	removeCcx := 0
	updates := 0
	roundUpdates := 0
	lastUpdate := 0

	channelData := make(chan chanData, CcxForceMaxUsers+5)

	roundTime.StartTime = time.Now()
	timeStart := time.Now()

	for i := 0; i < len(data); i++ {
		if data[i].isEnableForCcx() == operation {
			continue
		}
		if !data[i].isEnableForCcx() && operation {
			setCcx++
			channelData <- chanData{
				Id:      i,
				SetLine: true,
			}
			updates++
			roundUpdates++
		}
		if data[i].isEnableForCcx() && !operation {
			removeCcx++
			channelData <- chanData{
				Id:      i,
				SetLine: false,
			}
			updates++
			roundUpdates++
		}
		if needUpdate-removeCcx-setCcx < 1 {
			break
		}

		if roundUpdates >= CcxForceMaxUsers {
			if updates == lastUpdate {
				log.Errorf("same updates numbers")
				break
			}
			log.Infof("start run GO routine for manipulate with AXL %d for %d updates", updates/CcxForceMaxUsers, CcxForceMaxUsers)
			runUpdateBatch(&channelData, updates/CcxForceMaxUsers, &timeStart)
			lastUpdate = updates
			roundUpdates = 0
			if (updates/CcxForceMaxUsers)%5 == 0 {
				fmt.Print(roundTime.StringFinal())
			}
		}
		// TODO: make only 2 rounds
		//if updates%(CcxForceMaxUsers*2) == 0 && updates > 0 {
		//	break
		//}
	}
	if len(channelData) > 0 {
		log.Infof("start run GO routine for manipulate with AXL last update for %d updates", updates%CcxForceMaxUsers)
		runUpdateBatch(&channelData, updates/CcxForceMaxUsers+1, &timeStart)
	}
	log.Infof("system set %d and remove %d agents", setCcx, removeCcx)
}

func runUpdateBatch(channelData *chan chanData, round int, startTime *time.Time) {
	var wg sync.WaitGroup
	updates := len(*channelData)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go processAxlStream(channelData, &wg)
	}
	wg.Wait()
	axlEnd := time.Now()
	var force *ccxForceResponse
	var err error
	for i := 0; i < CxxForceDownRepeat; i++ {
		force, err = CcxResourceForceSync()
		if err == nil {
			break
		}
		if i < 2 && err.Error() == "CCX service is down" {
			log.Infof("CCX service is down - wait additional %d minutes", CcxForceDownTime)
			time.Sleep(CcxForceDownTime * time.Minute)
			continue
		}
		break
	}
	ccxEnd := time.Now()
	roundTime.addUpdateRound(round, updates, *startTime, axlEnd, ccxEnd)
	if err != nil {
		log.Errorf("round %d has problem force update CCX users, %s", round, err)
		programExit(1)
	}
	log.Infof("round %d after force update get %d/%d agents", round, len(force.AgentInfo), len(force.getGenerated()))
	time.Sleep(CcxForceWaitTime * time.Second)
	*startTime = time.Now()
}

func processAxlStream(updateChannel *chan chanData, wd *sync.WaitGroup) {
	id := fmt.Sprintf("ASYNC-%s%s", AxlIdPrefix, RandomString())
	log.WithField("id", id).Tracef("start async procedure %s", id)
	data := axlEndUsersList.getGeneratedUsers()
	var wg sync.WaitGroup
	for {
		select {
		case d := <-*updateChannel:
			wg.Add(1)
			go data[d.Id].switchCcxLine(axlServer, d.SetLine, &wg)
			wg.Wait()
		default:
			wd.Done()
			log.WithField("id", id).Tracef("exit async procedure %s", id)
			return
		}
	}
}

func correctTeams() {
	log.Debugf("correct teams from [%d] to [%d]", len(ccxTeamActiveList.getGeneratedTeams()), ccxUserActiveList.getNecessaryTeams())
}

func processCommands() {
	ccxApiServer = newCcxServer()
	axlServer = newAxlServer()

	readActualData()
	if *showOnly {
		return
	}

	if *Config.finalNumber == 0 {
		if len(axlEndUsersList.getGeneratedCcxUsers()) == 0 {
			log.Infof("program does nothing all generated users on server [%s] are removed", *Config.ccServer)
			return
		}
		log.Infof("program delete all generated users on server [%s]", *Config.ccServer)
		setAgents()
		correctTeams()
	} else {
		if len(ccxUserActiveList.Resource) == *Config.finalNumber {
			log.Infof("program does nothing enabled users on server [%s] is [%d]", *Config.ccServer, len(ccxUserActiveList.Resource))
			return
		}
		log.Infof("program set number of users from [%d] to [%d] on server [%s]", len(ccxUserActiveList.Resource), *Config.finalNumber, *Config.ccServer)
		setAgents()
		correctTeams()

	}
}

func main() {
	timeStart := time.Now()
	exitCode := 0

	kingpin.Version(VersionDetail())
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	log.SetLevel(getLevel())
	roundTime.StartTime = time.Now()

	// start process data
	if Config.validate() {
		fmt.Println(Config.toString())
		processCommands()
	}
	timeDuration := time.Since(timeStart)
	log.Infof("program run [%s]", timeDuration.String())
	programExit(exitCode)
}
