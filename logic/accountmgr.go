package logic

import (
	"sync"
	"wentserver/config"
	wentproto "wentserver/proto"
)

type AccountManager struct {
	AccountInfos wentproto.AccountInfos
	lock         *sync.RWMutex
}

func NewAccountManager(data wentproto.AccountInfos) (*AccountManager, error) {
	return &AccountManager{
		AccountInfos: data,
		lock:         &sync.RWMutex{},
	}, nil
}

func (pm *AccountManager) GetAccount(name string) (interface{}, error) {
	pm.lock.RLock()
	defer pm.lock.RUnlock()
	infomap := pm.AccountInfos.GetAccountmap()
	if infomap == nil {
		return nil, config.ErrAccountMapEmpty
	}

	accountInfo, ok := infomap[name]
	if !ok {
		return nil, config.ErrAccountNameNotExist
	}
	return *accountInfo, nil
}

func (am *AccountManager) RegAccount(name string, act *wentproto.AccountInfo) {
	am.lock.Lock()
	defer am.lock.Unlock()
	infomap := am.AccountInfos.GetAccountmap()
	infomap[name] = act
}
