package main

import (
	"wentserver/config"
	"wentserver/log"
	"wentserver/logic"
	"wentserver/netmodel"
	"wentserver/wentdb"
)

func InitMgr() error {
	mgr := logic.GetAccountManagerIns()
	if mgr == nil {
		return config.ErrAccountMgrInit
	}

	basemgr := logic.GetPlayerManagerIns()
	if basemgr == nil {
		return config.ErrPlayerMgrInit
	}

	genuidmgr := logic.GetGenuidIns()
	if genuidmgr == nil {
		return config.ErrGenuidMgrFailed
	}
	return nil
}

func main() {
	logic.RegServerHandlers()
	logins := log.InitLog("./server.log")
	if logins == nil {
		return
	}
	defer logins.CloseLogMgr()
	dh := wentdb.GetDBHandlerIns()
	if dh == nil {
		return
	}
	dh.StartSaveRoutine()
	defer dh.CloseDB()
	err := InitMgr()
	if err != nil {
		return
	}
	wt, err := netmodel.NewServer()
	if err != nil {
		return
	}
	defer wt.Close()
	wt.AcceptLoop()
}
