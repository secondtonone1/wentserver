package wentdb

import (
	"wentserver/config"

	"github.com/syndtr/goleveldb/leveldb"
)

type DBManager struct {
	db *leveldb.DB
}

func (dbm *DBManager) InitDB(path string) error {
	var err1 error
	dbm.db, err1 = leveldb.OpenFile(path, nil)
	if err1 != nil {
		return config.ErrDBInitFailed
	}
	return nil
}

func (dbm *DBManager) CloseDB() {
	dbm.db.Close()
}

func (dbm *DBManager) GetData(key []byte) ([]byte, error) {
	// 读取某条数据
	data, err := dbm.db.Get(key, nil)
	if err != nil {
		return nil, config.ErrDBGetValueFailed
	}
	return data, nil
}

func (dbm *DBManager) PutData(key []byte, value []byte) error {
	// 读取某条数据
	err := dbm.db.Put(key, value, nil)
	if err != nil {
		return config.ErrDBPutValueFailed
	}
	return nil
}
