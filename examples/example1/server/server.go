package main

import (
	"wentserver/logic"
	"wentserver/netmodel"
)

func main() {
	logic.RegServerHandlers()
	wt, err := netmodel.NewServer()
	if err != nil {
		return
	}
	defer wt.Close()
	wt.AcceptLoop()
}
