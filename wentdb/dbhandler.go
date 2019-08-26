package wentdb

import (
	"context"
	"fmt"
	"sync"
	"wentserver/config"
	log "wentserver/log"
)

type MsgSave struct {
	Key   []byte
	Value []byte
}

type DBHandler struct {
	dbm         *DBManager
	savechan    chan *MsgSave
	dbcontext   context.Context
	ctxcancle   context.CancelFunc
	dbwaitgroup *sync.WaitGroup
	savexit     chan struct{}
	lock        *sync.RWMutex
	bclosed     bool
}

func NewDBHandler() (*DBHandler, error) {
	dbm, err := InitDB(config.DB_PATH)

	if err != nil {
		return nil, err
	}

	dbhandler := new(DBHandler)
	dbhandler.dbm = dbm
	dbhandler.savechan = make(chan *MsgSave, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	dbhandler.dbcontext = ctx
	dbhandler.ctxcancle = cancel
	dbhandler.dbwaitgroup = &sync.WaitGroup{}
	dbhandler.savexit = make(chan struct{})
	dbhandler.lock = &sync.RWMutex{}
	dbhandler.bclosed = false
	dbhandler.dbwaitgroup.Add(config.SAVEROUTINE_COUNT)
	return dbhandler, nil
}

func (dh *DBHandler) CloseDB() {
	dh.lock.Lock()
	defer dh.lock.Unlock()
	onceDHClose.Do(func() {
		dh.ctxcancle()
		dh.bclosed = true
		if dh.dbm == nil {
			return
		}
		dh.dbm.CloseDB()
	})

}

var insDBHandler *DBHandler
var onceDBHandler sync.Once
var onceDHClose sync.Once

func GetDBHandlerIns() *DBHandler {
	onceDBHandler.Do(func() {
		var err error
		insDBHandler, err = NewDBHandler()
		if err != nil {
			insDBHandler = nil
		}
	})
	return insDBHandler
}

func (dh *DBHandler) LoadAccountData() [][]byte {
	return dh.dbm.LoadAccountData()
}

func (dh *DBHandler) LoadPlayerBaseData() [][]byte {
	return dh.dbm.LoadPlayerBaseData()
}

func (dh *DBHandler) LoadGenuid() []byte {
	return dh.dbm.LoadGenuid()
}

func (dh *DBHandler) PostMsgToSave(msg *MsgSave) error {
	dh.lock.RLock()
	defer dh.lock.RUnlock()
	select {
	case <-dh.savexit:
		fmt.Println("all save routines exited")
		log.GetLogManagerIns().Println("all save routines exited")
		return config.ErrAllSaveRoutineExit
	default:
		if dh.bclosed == true {
			fmt.Println("dbhandler main thread exit")
			return config.ErrDBHandlerExit
		}
		dh.savechan <- msg
		fmt.Println("msg post into the save chan")
		log.GetLogManagerIns().Println("msg post into the save chan")
		return nil
	}
}

func (dh *DBHandler) StartSaveRoutine() {
	go func(wg *sync.WaitGroup, savexit chan struct{}) {
		defer func(st chan struct{}) {
			fmt.Println("save watcher catches all save routine exit")
			log.GetLogManagerIns().Println("save watcher catches all save routine exit")
			close(st)
		}(savexit)
		wg.Wait()
	}(dh.dbwaitgroup, dh.savexit)

	for i := 0; i < config.SAVEROUTINE_COUNT; i++ {
		go func(dbcontext context.Context, savechan chan *MsgSave, wg *sync.WaitGroup, index int) {
			for {
				defer wg.Done()
				select {
				case <-dbcontext.Done():
					fmt.Println("dhhandler main thread exit")
					return
				case msg, ok := <-savechan:
					if !ok {
						fmt.Println("dhhandler main thread close savechan")
						return
					}
					fmt.Println("saveroutine index is : ", index, "save msg", msg)
					log.GetLogManagerIns().Println("saveroutine index is : ", index, "save msg", msg)
					err := dh.dbm.PutData(msg.Key, msg.Value)
					if err != nil {
						fmt.Println("saveroutine ", index, "save data failed")
						log.GetLogManagerIns().Println("saveroutine ", index, "save data failed")
						return
					}
				}
			}
		}(dh.dbcontext, dh.savechan, dh.dbwaitgroup, i)
	}

}
