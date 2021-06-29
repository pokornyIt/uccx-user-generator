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

const (
	ApplicationName = "uccx-user-generator"
	TimeFormat      = "15:04:05.0000"           // time format
	DateTimeFormat  = "2006-01-02 15:04:05.000" // Full date time format
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

func processCommands() {
	if *Config.finalNumber == 0 {
		log.Infof("program delete all generated users on server [%s]", *Config.ccServer)
	} else {
		log.Infof("program correct number of users to [%d] on server [%s]", *Config.finalNumber, *Config.ccServer)
	}

	ccxServer := newCcxServer()
	axlServer := newAxlServer()
	var wg sync.WaitGroup

	fmt.Printf(".... start\r\n")

	wg.Add(1)
	go asyncCcxResourceList(ccxServer, &wg)
	fmt.Printf(".... go ccx user\r\n")

	wg.Add(1)
	go asyncAxlReadUsers(axlServer, &wg)
	fmt.Printf(".... go axl users\r\n")

	wg.Add(1)
	go asyncCcxTeamList(ccxServer, &wg)
	fmt.Printf(".... go ccx team\r\n")

	wg.Wait()

	log.Debugf("finish read %d CCX teams", len(ccxTeamActiveList.Team))
	log.Debugf("finish read %d CCX users", len(ccxUserActiveList.Resource))
	log.Debugf("finish read %d CUCM users", len(axlEndUsersList.Rows))

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
