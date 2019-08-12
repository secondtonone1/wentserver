package logic

import (
	"sync"
	"wentserver/config"
	wentproto "wentserver/proto"
)

type PlayerManager struct {
	PlayerInfos wentproto.PlayerInfos
	lock        *sync.RWMutex
}

func NewPlayerManager(data wentproto.PlayerInfos) (*PlayerManager, error) {
	return &PlayerManager{
		PlayerInfos: data,
		lock:        &sync.RWMutex{},
	}, nil
}

func (pm *PlayerManager) GetPlayerById(id int64) (interface{}, error) {
	pm.lock.RLock()
	defer pm.lock.RUnlock()
	infomap := pm.PlayerInfos.GetPlayerinfomap()
	if infomap == nil {
		return nil, config.ErrPlayerMapEmpty
	}

}
