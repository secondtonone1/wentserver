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

func newAccountManager(data [][]byte) (*AccountManager, error) {

	am := new(AccountManager)
	for _, value := range data {
		acinfo := &wentproto.AccountInfo{}
		err := proto.Unmarshal(value, acinfo)
		if err != nil {
			continue
		}
		am.AccountInfos[acinfo.Accountname] = acinfo
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

var accins *AccountManager
var acconce sync.Once
var err error

func GetAccountManagerIns() *AccountManager {
	acconce.Do(func() {
		accins, err = newAccountManager(wentdb.GetDBManagerIns().LoadAccountData())
	})
	return accins
}
