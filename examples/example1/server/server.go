package main

import (
	"wentserver/config"
	"wentserver/logic"
	"wentserver/netmodel"
	"wentserver/wentdb"
)

func InitDB() (*wentdb.DBManager, error) {
	var dbmgr *wentdb.DBManager = wentdb.GetDBManagerIns()
	err := dbmgr.InitDB("./lvdb")
	if err != nil {
		return nil, config.ErrDBInitFailed
	}
	return dbmgr, nil
}

func InitMgr() error {
	mgr := logic.GetAccountManagerIns()
	if mgr == nil {
		return config.ErrAccountMgrInit
	}

	basemgr := logic.GetPlayerManagerIns()
	if basemgr == nil {
		return config.ErrPlayerMgrInit
	}

	return nil
}

func main() {
	logic.RegServerHandlers()

	dbmgr, err := InitDB()
	if err != nil {
		return
	}
	defer dbmgr.CloseDB()
	err = InitMgr()
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
