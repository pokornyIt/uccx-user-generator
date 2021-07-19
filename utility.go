package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // map for random string
	letterIdxBits = 6                                                      // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits                                     // # of letter indices fitting in 63 bits
)

func RandomString() string {
	sb := strings.Builder{}
	sb.Grow(maxRandomSize)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := maxRandomSize-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func isValidServer(server string, name string) error {
	match, err := regexp.MatchString("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$", server)
	log.Tracef("test if [%s] is valid IP for %s", server, name)
	if !match || err != nil {
		match, err = regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", server)
		log.Tracef("test if [%s] is valid FQDN %s", server, name)
		if !match || err != nil {
			return fmt.Errorf("defined %s server address isn't valid FQDN or IP address", name)
		}
	}
	return nil
}

func appVersion() string {
	if len(Version) > 0 {
		return fmt.Sprintf("%s/%s", ApplicationName, Version)
	}
	return fmt.Sprintf("%s/1.0", ApplicationName)
}

func programExit(code int) {
	a := Config.toString()

	if len(roundTime.RoundTime) > 0 {
		a = roundTime.String()
	}
	a = a + showStartInfo()
	fmt.Print(a)
	name := fmt.Sprintf("./log/_%s_end.log", roundTime.StartTime.Format(DateTimeFileFormat))
	time.Sleep(1 * time.Second)
	err := ioutil.WriteFile(name, []byte(a), 0644)
	if err != nil {
		log.Errorf("final file name [%s] - %s", name, err)
		time.Sleep(1 * time.Second)
		name = fmt.Sprintf("./_%s_end.log", roundTime.StartTime.Format(DateTimeFileFormat))
		_ = ioutil.WriteFile(name, []byte(a), 0644)
	}
	time.Sleep(100 * time.Millisecond)
	os.Exit(code)
}

func showStartInfo() string {
	allTime := roundTime.getAllTime()
	updates := roundTime.getUpdates()
	dur := allTime.String()
	l := 6
	if len(dur) > l {
		l = len(dur) - 1
	} else {
		for {
			if len(dur) > l {
				break
			}
			dur = " " + dur
		}
	}
	format := fmt.Sprintf("%%s%%-20s%%%dd\r\n", l)
	formatStr := fmt.Sprintf("%%s%%-20s%%%ds\r\n", l)
	formatTime := fmt.Sprintf("%%s%%-20s%%%ds\r\n", l+1)
	a := "\r\nStart data on CUCM:\r\n"
	a = fmt.Sprintf(format, a, "All users", len(axlEndUsersList.Rows))
	a = fmt.Sprintf(format, a, "Generated users", len(axlEndUsersList.getGeneratedUsers()))
	a = fmt.Sprintf(format, a, "Test enabled users", len(axlEndUsersList.getGeneratedCcxUsers()))
	a = fmt.Sprintf("%s\r\nStart data on CCX:\r\n", a)
	a = fmt.Sprintf(format, a, "All users", len(ccxUserActiveList.Resource))
	a = fmt.Sprintf(format, a, "Generated users", len(ccxUserActiveList.getGeneratedUsers()))
	a = fmt.Sprintf(format, a, "All teams", len(ccxTeamActiveList.Team))
	a = fmt.Sprintf(format, a, "Generated teams", len(ccxTeamActiveList.getGeneratedTeams()))
	a = fmt.Sprintf("%s\r\nProcess data:\r\n", a)
	a = fmt.Sprintf(format, a, "After force wait", CcxForceWaitTime)
	if roundTime.Direction {
		a = fmt.Sprintf(formatStr, a, "Direction", "add")
	} else {
		a = fmt.Sprintf(formatStr, a, "Direction", "remove")
	}
	a = fmt.Sprintf(format, a, "Bucket size", CcxForceMaxUsers)
	a = fmt.Sprintf(format, a, "Updates", updates)
	a = fmt.Sprintf(formatTime, a, "Duration", dur)
	if updates == 0 {
		updates = 1
	}
	t := time.Duration(allTime.Nanoseconds()/int64(updates)) * time.Nanosecond
	a = fmt.Sprintf(formatTime, a, "Round time", t.String())
	return a
}
