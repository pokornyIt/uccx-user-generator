package main

import (
	"fmt"
	"strconv"
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
	StartTime time.Time
	RoundTime []roundTimeStatData
}

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
}

func (s *statData) String() string {
	formatString := "%s    %5s / %15s / %15s / %15s / %15s / %15s\r\n"
	a := fmt.Sprintf(formatString, "\r\n", "Round", "Axl Duration", "CCX Duration", "Sleep Duration", "Round Duration", "Updates")
	axlDur := time.Duration(0)
	ccxDur := time.Duration(0)
	sleepDur := time.Duration(0)
	allDur := time.Duration(0)
	allUpdates := 0
	for _, data := range s.RoundTime {
		a = fmt.Sprintf(formatString, a, strconv.Itoa(data.RoundId), data.AxlDuration.String(), data.CcxDuration.String(),
			data.SleepDuration.String(), data.RoundDuration.String(), strconv.Itoa(data.Updates))
		axlDur = axlDur + data.AxlDuration
		ccxDur = ccxDur + data.CcxDuration
		sleepDur = sleepDur + data.SleepDuration
		allDur = allDur + data.RoundDuration
		allUpdates = allUpdates + data.Updates
	}
	a = fmt.Sprintf(formatString, a, "sum", axlDur, ccxDur, sleepDur, allDur, strconv.Itoa(allUpdates))
	return a
}
