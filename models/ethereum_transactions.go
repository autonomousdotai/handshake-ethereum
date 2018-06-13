package models

import (
	_ "encoding/gob"
	"time"

	_ "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type EthereumTransactions struct {
	DateCreated  time.Time
	DateModified time.Time
	ID           int64
	UserID       int64
	Network      string `json:"network"`
	ChainID      int64  `json:"chain_id"`
	Contract     string `json:"contract"`
	RefType      string `json:"ref_type"`
	RefID        int64  `json:"ref_id"`
	Hash         string `json:"hash"`
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
