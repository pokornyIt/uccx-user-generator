package main

import (
	"regexp"
	"sync"
)

var (
	ccxUserActiveList  *ccxResourceList // CCX Active User list
	ccxUserActiveMutex sync.Mutex       // for async data update

	ccxTeamActiveList *ccxTeamList                           // CCX Team list
	ccxTeamListMutex  sync.Mutex                             // for async data update
	ccxTeamRegex      = regexp.MustCompile(CcxTeamNameRegex) // regex for validate teams

	axlEndUsersList  *axlEndUserList                        // AXL User list
	axlEndUsersMutex sync.Mutex                             // for async data update
	axlUserRegex     = regexp.MustCompile(CcxUserNameRegex) // regex for validate AXL/ccx user name
)
