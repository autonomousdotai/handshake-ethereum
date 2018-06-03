package controller

import (
	"log"
	"encoding/json"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"path/filepath"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
	"github.com/ethereum/go-ethereum"
	"context"
	"reflect"
	"github.com/autonomousdotai/handshake-ethereum/models"
	"github.com/autonomousdotai/handshake-ethereum/dao"
	"math/big"
	"google.golang.org/api/option"
	"cloud.google.com/go/pubsub"
	"github.com/autonomousdotai/handshake-ethereum/param"
	"github.com/pkg/errors"
)

var ethereumLogsDao = dao.EthereumLogsDao{}

type Controller struct {
	Processers []*Processer
}

type Processer struct {
	Agr         param.Agr
	Client      *ethclient.Client
	Addresses   []common.Address
	Topics      []string
	Abi         abi.ABI
	PubsubTopic *pubsub.Topic
}

func NewConcotrller(agrs []param.Agr) (*Controller, error) {
	controller := Controller{}
	opt := option.WithCredentialsFile(param.Conf.CredsFile)
	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, param.Conf.ProjectId, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, agr := range agrs {
		processer, err := NewProcesser(agr, pubsubClient)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		controller.Processers = append(controller.Processers, processer)
	}
	return &controller, nil
}

func (controller *Controller) Process() {
	for _, processer := range controller.Processers {
		go processer.Process()
	}
}

func NewProcesser(agr param.Agr, pubsubClient *pubsub.Client) (*Processer, error) {
	processer := Processer{}
	processer.Agr = agr
	client, err := ethclient.Dial(agr.ChainNetwork)
	if err != nil {
		log.Println("error:", err)
		return nil, err
	}
	processer.Client = client
	processer.Addresses = []common.Address{
		common.HexToAddress(agr.ContractAddress),
	}
	path, err := filepath.Abs(param.ABI_FILES[agr.Contract])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	abiIns, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	processer.Abi = abiIns

	pubsubTopic, err := pubsubClient.CreateTopic(context.Background(), agr.PubsubName)
	if err != nil {
		log.Println(err)
	} else {
		processer.PubsubTopic = pubsubTopic
	}

	return &processer, nil
}

func (processer *Processer) Process() (error) {
	log.Println("contract address", processer.Agr.ContractAddress)
	for _, event := range processer.Agr.Events {
		log.Println("process for event ", event)

		contractLogs := ethereumLogsDao.GetByFilter(processer.Agr.ContractAddress, event)
		q := ethereum.FilterQuery{}
		q.Addresses = processer.Addresses
		q.FromBlock = nil
		if contractLogs.ID > 0 {
			q.FromBlock = big.NewInt(contractLogs.BlockNumber + 1)
		}
		q.ToBlock = nil
		q.Topics = [][]common.Hash{[]common.Hash{processer.Abi.Events[event].Id()}}
		logs, err := processer.Client.FilterLogs(context.Background(), q)
		if err != nil {
			log.Println(err)
			return err
		}
		abiStructs := param.ABI_STRUCTS[processer.Agr.Contract]
		for _, logE := range logs {
			hash := logE.TxHash.String()
			val, ok := abiStructs[event]
			if !ok {
				log.Println(errors.New("event " + event + " struct is missed"))
				break
			}
			outptr := reflect.New(reflect.TypeOf(val))
			err = processer.Abi.Unpack(outptr.Interface(), event, logE.Data)
			if err != nil {
				if err != nil {
					log.Println(err)
					break
				}
			} else {
				data, err := processer.MakeData(event, outptr.Interface())
				if err != nil {
					log.Println(err)
					break
				}
				go processer.SaveDB(processer.Agr.ChainId, processer.Agr.ContractAddress, event, int64(logE.BlockNumber), int64(logE.Index), hash, data)
				go processer.PubSub(processer.Agr.ChainId, processer.Agr.ContractAddress, event, int64(logE.BlockNumber), int64(logE.Index), hash, data)
			}
		}
	}
	return nil
}

func (processer *Processer) MakeData(event string, source interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	jsonStr, err := json.Marshal(source)
	if err != nil {
		log.Println(err)
		return result, err
	}
	err = json.Unmarshal(jsonStr, &result)
	if err != nil {
		log.Println(err)
		return result, err
	}

	for k, v := range result {
		switch v.(type) {
		default:
			log.Println("unexpected type %T", v)
			break
		case float64:
			result[k] = int64(v.(float64))
			break
		case []interface{}:
			str := ""
			for _, i := range v.([]interface{}) {
				switch i.(type) {
				case float64:
					str += string([]byte{byte(i.(float64))})
					break
				}
			}
			str = strings.Trim(str, string([]byte{0}))
			result[k] = str
			break
		}
	}

	return result, nil
}

func (processer *Processer) SaveDB(chainId int, address string, event string, blockNumber int64, logIndex int64, hash string, data map[string]interface{}) (error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		log.Println("json.Marshal(data)", err)
		return err
	}
	ethereumLogs := models.EthereumLogs{}
	ethereumLogs.ChainId = chainId
	ethereumLogs.Address = address
	ethereumLogs.Event = event
	ethereumLogs.BlockNumber = blockNumber
	ethereumLogs.LogIndex = logIndex
	ethereumLogs.Hash = hash
	ethereumLogs.Data = string(jsonStr)

	ethereumLogs, err = ethereumLogsDao.Create(ethereumLogs, nil)
	if err != nil {
		log.Println("ethereumLogsDao.Create", err)
		return err
	}

	return nil
}

func (processer *Processer) PubSub(chainId int, address string, event string, blockNumber int64, logIndex int64, hash string, data map[string]interface{}) (error) {
	pubsubData := map[string]interface{}{}
	pubsubData["address"] = address
	pubsubData["event"] = event
	pubsubData["block_number"] = blockNumber
	pubsubData["log_index"] = logIndex
	pubsubData["hash"] = hash
	pubsubData["data"] = data
	jsonStr, err := json.Marshal(pubsubData)
	if err != nil {
		log.Println("json.Marshal(pubsubData)", err)
		return err
	}
	log.Println(string(jsonStr))
	if processer.PubsubTopic != nil {
		res := processer.PubsubTopic.Publish(context.Background(), &pubsub.Message{Data: jsonStr})
		log.Println(res)
	}
	return nil
}
