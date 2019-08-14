package main

import (
	"wentserver/logic"
	"wentserver/netmodel"
	"wentserver/wentdb"
)

func main() {
	logic.RegServerHandlers()
	var dbmgr *wentdb.DBManager = wentdb.GetDBManagerIns()
	err := dbmgr.InitDB("./lvdb")
	if err != nil {
		return
	}
	defer dbmgr.CloseDB()
	wt, err := netmodel.NewServer()
	if err != nil {
		return
	}
	defer wt.Close()
	wt.AcceptLoop()
}
