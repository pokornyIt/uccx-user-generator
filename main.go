package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"os"
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
	GoVersion    = runtime.Version() // build GO version
	infoTmpl     = `
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

func setAgents(many int) {
	data := axlEndUsersList.getGeneratedUsers()
	setCcx := 0
	removeCcx := 0
	var wg sync.WaitGroup

	channelData := make(chan chanData, 10005)

	// TODO: need process per evey 100 changes no more. 1955 changes make CCX force refresh more than 60 minutes
	for i := 0; i < len(data); i++ {
		if i < many {
			if !data[i].isEnableForCcx() {
				setCcx++
				//wg.Add(1)
				//go data[i].switchCcxLine(axlServer, true, &wg)
				channelData <- chanData{
					Id:      i,
					SetLine: true,
				}
			}
		} else {
			if data[i].isEnableForCcx() {
				removeCcx++
				//wg.Add(1)
				//go data[i].switchCcxLine(axlServer, false, &wg)
				channelData <- chanData{
					Id:      i,
					SetLine: false,
				}
			}
		}
	}

	log.Infof("start run GO routine for manipulate with AXL")
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go processAxlStream(&channelData, &wg)
	}
	wg.Wait()
	log.Infof("system set %d and remove %d agents ", setCcx, removeCcx)
}

func processAxlStream(updateChannel *chan chanData, wd *sync.WaitGroup) {
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

	// fro start read actual data
	readActualData()

	if *Config.finalNumber == 0 {
		if len(axlEndUsersList.getGeneratedCcxUsers()) == 0 {
			log.Infof("program does nothing all generated users on server [%s] are removed", *Config.ccServer)
		} else {
			log.Infof("program delete all generated users on server [%s]", *Config.ccServer)
			setAgents(0)
			correctTeams()
		}
	} else {
		log.Infof("program decrease number of users from [%d] to [%d] on server [%s]", len(ccxUserActiveList.Resource), *Config.finalNumber, *Config.ccServer)
		setAgents(axlEndUsersList.getNeedEnabledUsers())
		correctTeams()
	}

	force, err := CcxResourceForceSync()
	if err != nil {
		log.Errorf("problem force update CCX users, %s", err)
	} else {
		log.Infof("after force update get %d/%d agents", len(force.AgentInfo), len(force.getGenerated()))
	}
}

func main() {
	timeStart := time.Now()
	exitCode := 0

	kingpin.Version(VersionDetail())
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	log.SetLevel(getLevel())

	// start process data
	if Config.validate() {
		fmt.Println(Config.toString())
		processCommands()
	}
	timeDuration := time.Since(timeStart)
	log.Infof("program run [%s]", timeDuration.String())
	time.Sleep(time.Second)
	os.Exit(exitCode)

}
