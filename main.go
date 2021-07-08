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
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
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
	ccxServer := newCcxServer()
	axlServer := newAxlServer()
	var wg sync.WaitGroup

	wg.Add(1)
	go asyncCcxResourceList(ccxServer, &wg)

	wg.Add(1)
	go asyncAxlReadUsers(axlServer, &wg)

	wg.Add(1)
	go asyncCcxTeamList(ccxServer, &wg)

	wg.Wait()
	log.Infof("Finish initial data read from all sources")

	log.Debugf("finish read %d/%d CCX teams", len(ccxTeamActiveList.Team), len(ccxTeamActiveList.getGeneratedTeams()))
	log.Debugf("finish read %d/%d CCX users", len(ccxUserActiveList.Resource), len(ccxUserActiveList.getGeneratedUsers()))
	log.Debugf("finish read %d/%d CUCM users", len(axlEndUsersList.Rows), len(axlEndUsersList.getGeneratedUsers()))
}

func disableAgents(many int) {}

func enableAgents(many int) {}

func correctTeams() {
	log.Debugf("correct teams from [%d] to [%d]", len(ccxTeamActiveList.getGeneratedTeams()), ccxUserActiveList.getNecessaryTeams())
}

func processCommands() {
	// fro start read actual data
	readActualData()

	if *Config.finalNumber == 0 {
		if len(ccxUserActiveList.getGeneratedUsers()) == 0 {
			log.Infof("program does nothing all generated users on server [%s] are removed", *Config.ccServer)
		} else {
			log.Infof("program delete all generated users on server [%s]", *Config.ccServer)
			disableAgents(len(ccxUserActiveList.getGeneratedUsers()))
			correctTeams()
		}
	} else {
		if *Config.finalNumber < len(ccxUserActiveList.Resource) {
		} else if *Config.finalNumber < len(ccxUserActiveList.Resource) {
			log.Infof("program decrease number of users from [%d] to [%d] on server [%s]", len(ccxUserActiveList.Resource), *Config.finalNumber, *Config.ccServer)
			dif := len(ccxUserActiveList.Resource) - *Config.finalNumber
			disableAgents(dif)
			correctTeams()
		} else if *Config.finalNumber > len(ccxUserActiveList.Resource) {
			log.Infof("program increase number of users from [%d] to [%d] on server [%s]", len(ccxUserActiveList.Resource), *Config.finalNumber, *Config.ccServer)
			dif := *Config.finalNumber - len(ccxUserActiveList.Resource)
			enableAgents(dif)
			correctTeams()
		} else {
			log.Infof("program does nothing on server [%s] is necessary users [%d", *Config.ccServer, *Config.finalNumber)
			correctTeams()
		}
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
