package param

import (
	"math/big"
	"os"
	"log"
	"encoding/json"
)

var Conf Config

var PAYABLE = "payable"
var CROWDSALE = "crowdsale"
var CRYPTOSIGN = "cryptosign"
var STRUCTS = map[string]map[string]interface{}{}

func Initialize(confFile string) {
	file, err := os.Open(confFile)
	if err != nil {
		log.Println(err)
	}
	decoder := json.NewDecoder(file)
	conf := Config{}
	err = decoder.Decode(&conf)
	if err != nil {
		log.Println(err)
	}
	Conf = conf

	STRUCTS[PAYABLE] = map[string]interface{}{}
	STRUCTS[CROWDSALE] = map[string]interface{}{}
	STRUCTS[CRYPTOSIGN] = map[string]interface{}{}

	//crowdsale
	STRUCTS[CROWDSALE]["__init"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__shake"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Balance  *big.Int `json:"balance"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__unshake"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Balance  *big.Int `json:"balance"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__cancel"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__stop"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__refund"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	STRUCTS[CROWDSALE]["__withdraw"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
}

type Config struct {
	DbUrl     string `json:"db_url"`
	CredsFile string `json:"creds_file"`
	ProjectId string `json:"project_id"`
	Agrs      []Agr  `json:"agrs"`
}

type Agr struct {
	ChainId         int      `json:"chain_id"`
	ChainNetwork    string   `json:"chain_network"`
	AbiFile         string   `json:"abi_file"`
	Contract        string   `json:"contract"`
	ContractAddress string   `json:"contract_address"`
	Events          []string `json:"events"`
	PubsubName      string   `json:"pubsub_name"`
}
