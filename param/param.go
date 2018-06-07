package param

import (
	"encoding/json"
	"log"
	"math/big"
	"os"
)

var Conf Config

var CONTRACT_PAYABLE = "payable"
var CONTRACT_CROWDSALE = "crowdsale"
var CONTRACT_CRYPTOSIGN = "cryptosign"
var ABI_STRUCTS = map[string]map[string]interface{}{}
var ABI_FILES = map[string]string{}

func Initialize(confFile string) error {
	file, err := os.Open(confFile)
	if err != nil {
		log.Println(err)
		return err
	}
	decoder := json.NewDecoder(file)
	conf := Config{}
	err = decoder.Decode(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	Conf = conf

	ABI_FILES[CONTRACT_PAYABLE] = "./abi/payable.abi"
	ABI_FILES[CONTRACT_CROWDSALE] = "./abi/crowdsale.abi"
	ABI_FILES[CONTRACT_CRYPTOSIGN] = "./abi/cryptosign.abi"

	ABI_STRUCTS[CONTRACT_PAYABLE] = map[string]interface{}{}
	ABI_STRUCTS[CONTRACT_CROWDSALE] = map[string]interface{}{}
	ABI_STRUCTS[CONTRACT_CRYPTOSIGN] = map[string]interface{}{}

	//crowdsale conf
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__init"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__shake"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Balance  *big.Int `json:"balance"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__unshake"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Balance  *big.Int `json:"balance"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__cancel"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__stop"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__refund"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_CROWDSALE]["__withdraw"] = struct {
		Hid      *big.Int `json:"hid"`
		State    uint8    `json:"state"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		1,
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	//payable conf
	ABI_STRUCTS[CONTRACT_PAYABLE]["__init"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__shake"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__deliver"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__cancel"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__reject"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__accept"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	ABI_STRUCTS[CONTRACT_PAYABLE]["__withdraw"] = struct {
		Hid      *big.Int `json:"hid"`
		Offchain [32]byte `json:"offchain"`
	}{big.NewInt(1),
		[32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	return nil
}

type Config struct {
	DbUrl             string `json:"db_url"`
	CredsFile         string `json:"creds_file"`
	ProjectId         string `json:"project_id"`
	Agrs              []Agr  `json:"agrs"`
	RinkebyNetwork    string `json:"rinkeby_network"`
	RinkebyPrivateKey string `json:"rinkeby_private_key"`
}

type Agr struct {
	ChainId         int    `json:"chain_id"`
	ChainNetwork    string `json:"chain_network"`
	Contract        string `json:"contract"`
	ContractAddress string `json:"contract_address"`
	TopicName       string `json:"topic_name"`
}
