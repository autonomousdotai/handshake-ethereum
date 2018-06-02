package models

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm"
	_ "encoding/gob"
	"time"
)

type EthereumLogs struct {
	DateCreated  time.Time
	DateModified time.Time
	ID           int64
	ChainId      int
	Address      string
	Event        string
	BlockNumber  int64
	LogIndex     int64
	Hash         string
	Data         string
}

func (EthereumLogs) TableName() string {
	return "ethereum_logs"
}
