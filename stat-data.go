package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type roundTimeStatData struct {
	RoundId       int
	Updates       int
	CcxDuration   time.Duration
	AxlDuration   time.Duration
	RoundDuration time.Duration
	SleepDuration time.Duration
	RoundMessage  string
}

type statData struct {
	StartTime     time.Time
	ExpectUpdates int
	Direction     bool
	RoundTime     []roundTimeStatData
}

var statWriteDataMux sync.Mutex
var statWriteFileName = fmt.Sprintf("./log/_%s.log", time.Now().Format(DateTimeFileFormat))

func (s *statData) addUpdateRound(id int, updates int, startTime time.Time, axlEnd time.Time, ccxEnd time.Time) {
	endTime := time.Now()
	d := roundTimeStatData{
		RoundId:       id,
		Updates:       updates,
		CcxDuration:   ccxEnd.Sub(axlEnd),
		AxlDuration:   axlEnd.Sub(startTime),
		SleepDuration: endTime.Sub(ccxEnd),
		RoundDuration: endTime.Sub(startTime),
		RoundMessage:  fmt.Sprintf("start run GO routine for manipulate with round %d in time %s", id, endTime.Sub(startTime).String()),
	}
	s.RoundTime = append(s.RoundTime, d)
	go writeLog(s.StringFinal())
}

func writeLog(data string) {
	statWriteDataMux.Lock()
	_ = ioutil.WriteFile(statWriteFileName, []byte(data), 0644)
	statWriteDataMux.Unlock()
}

func (s *statData) String() string {
	a := fmt.Sprintf(StatFormatString, "\r\n", "Round", "Axl Duration", "CCX Duration", "Sleep Duration", "Round Duration", "Updates")
	axlDur := time.Duration(0)
	ccxDur := time.Duration(0)
	sleepDur := time.Duration(0)
	allDur := time.Duration(0)
	allUpdates := 0
	for _, data := range s.RoundTime {
		a = fmt.Sprintf(StatFormatString, a, strconv.Itoa(data.RoundId), data.AxlDuration.String(), data.CcxDuration.String(),
			data.SleepDuration.String(), data.RoundDuration.String(), strconv.Itoa(data.Updates))
		axlDur = axlDur + data.AxlDuration
		ccxDur = ccxDur + data.CcxDuration
		sleepDur = sleepDur + data.SleepDuration
		allDur = allDur + data.RoundDuration
		allUpdates = allUpdates + data.Updates
	}
	a = fmt.Sprintf(StatFormatString, a, "sum", axlDur, ccxDur, sleepDur, allDur, strconv.Itoa(allUpdates))
	return a
}

func (s *statData) StringFinal() string {
	axlDur := time.Duration(0)
	ccxDur := time.Duration(0)
	sleepDur := time.Duration(0)
	allDur := time.Duration(0)
	allUpdates := 0
	for _, data := range s.RoundTime {
		axlDur = axlDur + data.AxlDuration
		ccxDur = ccxDur + data.CcxDuration
		sleepDur = sleepDur + data.SleepDuration
		allDur = allDur + data.RoundDuration
		allUpdates = allUpdates + data.Updates
	}
	var a string
	if s.Direction {
		a = fmt.Sprintf("add %d", s.ExpectUpdates)
	} else {
		a = fmt.Sprintf("delete %d", s.ExpectUpdates)
	}

	return fmt.Sprintf(StatFormatString, "", a, axlDur, ccxDur, sleepDur, allDur, strconv.Itoa(allUpdates))
}

func (s *statData) getUpdates() int {
	allUpdates := 0
	for _, data := range s.RoundTime {
		allUpdates = allUpdates + data.Updates
	}
	return allUpdates
}

func (s *statData) getAllTime() time.Duration {
	allDur := time.Duration(0)
	for _, data := range s.RoundTime {
		allDur = allDur + data.RoundDuration
	}
	return allDur
}
