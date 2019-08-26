package logic

import (
	"sync"
	"wentserver/config"
	wentproto "wentserver/proto"
	"wentserver/wentdb"

	"github.com/gogo/protobuf/proto"
)

type UidGenerator struct {
	curgenuid int64
	lock      *sync.RWMutex
}

func newUidGenerator(uiddata []byte) (*UidGenerator, error) {
	ns := new(UidGenerator)
	ns.lock = &sync.RWMutex{}
	if uiddata == nil || len(uiddata) == 0 {
		ns.curgenuid = 1
		return ns, nil
	}
	uidinfo := &wentproto.GenerateUid{}
	err := proto.Unmarshal(uiddata, uidinfo)
	if err != nil {
		ns.curgenuid = 1
		return nil, config.ErrUidUnmarshFailed
	}
	ns.curgenuid = uidinfo.Generateuid
	return ns, nil
}

var gennuidins *UidGenerator
var genuidonce sync.Once

func GetGenuidIns() *UidGenerator {
	genuidonce.Do(func() {
		gennuidins, err = newUidGenerator(wentdb.GetDBManagerIns().LoadGenuid())
		if err != nil {
			gennuidins = nil
		}
	})
	return gennuidins
}

func (ug *UidGenerator) generateuid() (int64, error) {
	ug.lock.Lock()
	defer ug.lock.Unlock()
	rt := ug.curgenuid
	ug.curgenuid++
	uidinfo := &wentproto.GenerateUid{
		Generateuid: ug.curgenuid,
	}
	values, err := proto.Marshal(uidinfo)
	if err != nil {
		return -1, config.ErrGenuidFailed
	}

	wentdb.GetDBHandlerIns().PostMsgToSave(&wentdb.MsgSave{Key: []byte("genuid_"), Value: values})

	if err != nil {
		return -1, config.ErrGenuidFailed
	}
	return rt, nil
}
