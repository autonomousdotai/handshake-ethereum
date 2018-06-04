package models

import (
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm"
	_ "encoding/gob"
	"time"
)

type EthereumTransactions struct {
	DateCreated  time.Time
	DateModified time.Time
	ID           int64
	UserId       int64
	ChainId      int64
	ServiceName  string
	RefType      string
	RefId        int64
	Hash         string
	FromAddress  string
	ToAddress    string
	Gas          float64
	GasPrice     float64
	Value        float64
	Nonce        int
	Data         string
	GasUsed      float64
	Status       int
}

func (EthereumTransactions) TableName() string {
	return "ethereum_transactions"
}
