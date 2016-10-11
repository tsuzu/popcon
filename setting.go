package main

import (
	"sync"
)

type Setting struct {
	ReCAPTCHASite              string
	ReCAPTCHASecret            string
	CanCreateUser              bool
	CanCreateContestByNotAdmin bool
	DB                         string
	JudgeKey                   string
	ListeningEndpoint          string
}

type SettingManager struct {
	setting Setting
	mut     sync.RWMutex
}

func (sm *SettingManager) Set(setting Setting) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	sm.setting = setting
}

func (sm *SettingManager) Get() Setting {
	sm.mut.RLock()
	defer sm.mut.RUnlock()

	return sm.setting
}

var settingManager SettingManager
