package logic

import (
	"sync"
	"wentserver/config"
	wentproto "wentserver/proto"

	"wentserver/wentdb"

	"github.com/gogo/protobuf/proto"
)

type AccountManager struct {
	AccountInfos map[string]*wentproto.AccountInfo
	lock         *sync.RWMutex
}

func newAccountManager(data map[string]string) (*AccountManager, error) {

	am := new(AccountManager)
	for key, value := range data {
		acinfo := &wentproto.AccountInfo{}
		err := proto.Unmarshal([]byte(value), acinfo)
		if err != nil {
			continue
		}
		am.AccountInfos[key] = acinfo
	}
	am.lock = &sync.RWMutex{}
	return am, nil
}

func (pm *AccountManager) GetAccount(name string) (interface{}, error) {
	pm.lock.RLock()
	defer pm.lock.RUnlock()

	if pm.AccountInfos == nil {
		return nil, config.ErrAccountMapEmpty
	}

	accountInfo, ok := pm.AccountInfos[name]
	if !ok {
		return nil, config.ErrAccountNameNotExist
	}
	return accountInfo, nil
}

func (am *AccountManager) RegAccount(name string, act *wentproto.AccountInfo) {
	am.lock.Lock()
	defer am.lock.Unlock()
	am.AccountInfos[name] = act
}

var ins *AccountManager
var once sync.Once
var err error

func GetAccountManagerIns() *AccountManager {
	once.Do(func() {
		ins, err = newAccountManager(wentdb.GetDBManagerIns().LoadAccountData())
	})
	return ins
}
