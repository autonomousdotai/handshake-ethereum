package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"path/filepath"
	"reflect"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninjadotorg/handshake-ethereum/dao"
	"github.com/ninjadotorg/handshake-ethereum/models"
	"github.com/ninjadotorg/handshake-ethereum/param"
	"google.golang.org/api/option"
)

var (
	ethereumLogsDao         = dao.EthereumLogsDao{}
	ethereumTransactionsDao = dao.EthereumTransactionsDao{}
)

type Controller struct {
	LogsProcessers []*LogsProcesser
}

type LogsProcesser struct {
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
	pubsubClient, err := pubsub.NewClient(context.Background(), param.Conf.ProjectID, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, agr := range agrs {
		processer, err := NewLogsProcesser(agr, pubsubClient)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		controller.LogsProcessers = append(controller.LogsProcessers, processer)
	}
	return &controller, nil
}

func (controller *Controller) Process() {
	for _, processer := range controller.LogsProcessers {
		go processer.Process()
	}
}

func NewLogsProcesser(agr param.Agr, pubsubClient *pubsub.Client) (*LogsProcesser, error) {
	processer := LogsProcesser{}
	processer.Agr = agr
	client, err := ethclient.Dial(agr.ChainNetwork)
	if err != nil {
		log.Println("NewLogsProcesser", err)
		return nil, err
	}
	processer.Client = client
	processer.Addresses = []common.Address{
		common.HexToAddress(agr.ContractAddress),
	}
	path, err := filepath.Abs(param.ABI_FILES[agr.Contract])
	if err != nil {
		log.Println("NewLogsProcesser", err)
		return nil, err
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("NewLogsProcesser", err)
		return nil, err
	}
	abiIns, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		log.Println("NewLogsProcesser", err)
		return nil, err
	}
	processer.Abi = abiIns

	pubsubTopic := pubsubClient.Topic(agr.TopicName)
	existed, err := pubsubTopic.Exists(context.Background())

	if pubsubTopic == nil || !existed {
		pubsubTopic, err := pubsubClient.CreateTopic(context.Background(), agr.TopicName)
		if err != nil {
			log.Println("NewLogsProcesser", err)
		} else {
			processer.PubsubTopic = pubsubTopic
		}
	} else {
		processer.PubsubTopic = pubsubTopic
	}

	return &processer, nil
}

func (processer *LogsProcesser) Process() error {
	log.Println("contract address", processer.Agr.ContractAddress)
	for _, event := range processer.Abi.Events {
		log.Println("LogsProcesser.Process() for event ", event)

		contractLogs := ethereumLogsDao.GetByFilter(processer.Agr.ContractAddress, event.Name)
		q := ethereum.FilterQuery{}
		q.Addresses = processer.Addresses
		q.FromBlock = nil
		if contractLogs.ID > 0 {
			q.FromBlock = big.NewInt(contractLogs.BlockNumber + 1)
		}
		q.ToBlock = nil
		q.Topics = [][]common.Hash{[]common.Hash{processer.Abi.Events[event.Name].Id()}}
		etherLogs, err := processer.Client.FilterLogs(context.Background(), q)
		if err != nil {
			log.Println("LogsProcesser.Process()", err)
			return err
		}
		abiStructs := param.ABI_STRUCTS[processer.Agr.Contract]
		for _, etherLog := range etherLogs {
			hash := etherLog.TxHash.String()

			val, ok := abiStructs[event.Name]
			if !ok {
				log.Println(errors.New("event " + event.Name + " struct is missed"))
				break
			}
			outptr := reflect.New(reflect.TypeOf(val))
			err = processer.Abi.Unpack(outptr.Interface(), event.Name, etherLog.Data)
			if err != nil {
				if err != nil {
					log.Println("LogsProcesser.Process()", err)
					break
				}
			} else {
				data, err := processer.MigrateData(event.Name, outptr.Interface())
				if err != nil {
					log.Println("LogsProcesser.Process()", err)
					break
				}
				go processer.ProcessMessage(processer.Agr.ChainID, processer.Agr.ContractAddress, event.Name, int64(etherLog.BlockNumber), int64(etherLog.Index), hash, data)
			}
		}
	}
	return nil
}

func (processer *LogsProcesser) MigrateData(event string, source interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	jsonStr, err := json.Marshal(source)
	if err != nil {
		log.Println("LogsProcesser.MigrateData()", err)
		return result, err
	}
	err = json.Unmarshal(jsonStr, &result)
	if err != nil {
		log.Println("LogsProcesser.MigrateData()", err)
		return result, err
	}

	for k, v := range result {
		switch v.(type) {
		default:
			log.Println("LogsProcesser.MigrateData() unexpected type pos 1 %T", v)
			break
		case float64:
			result[k] = int64(v.(float64))
			break
		case []interface{}:
			str := ""
			for _, i := range v.([]interface{}) {
				switch i.(type) {
				default:
					log.Println("LogsProcesser.MigrateData() unexpected type pos 2 %T", v)
					break
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

func (processer *LogsProcesser) ProcessMessage(chainId int, contractAddress string, event string, blockNumber int64, logIndex int64, hash string, data map[string]interface{}) error {
	fromAddress := ""
	transaction, _, err := processer.Client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		log.Println("LogsProcesser.ProcessMessage()", err)
	} else {
		fromAddress = transaction.From().String()
	}
	ethereumLogs, err := processer.SaveDB(processer.Agr.ChainID, fromAddress, processer.Agr.ContractAddress, event, blockNumber, logIndex, hash, data)
	if err != nil {
		log.Println("LogsProcesser.ProcessMessage()", err)
	}
	res, err := processer.PubSub(processer.Agr.ChainID, fromAddress, processer.Agr.ContractAddress, event, blockNumber, logIndex, hash, data)
	if err != nil {
		log.Println("LogsProcesser.ProcessMessage()", err)
		return err
	}
	if res != nil && ethereumLogs.ID > 0 {
		serverID, err := res.Get(context.Background())
		if err != nil {
			log.Println("LogsProcesser.ProcessMessage()", err)
		}
		ethereumLogs.PubsubMsgId = serverID
		ethereumLogs, err = ethereumLogsDao.Update(ethereumLogs, nil)
		if err != nil {
			log.Println("LogsProcesser.ProcessMessage()", err)
			return nil
		}
	}
	return nil
}

func (processer *LogsProcesser) SaveDB(chainId int, fromAddress string, contractAddress string, event string, blockNumber int64, logIndex int64, hash string, data map[string]interface{}) (models.EthereumLogs, error) {
	ethereumLogs := models.EthereumLogs{}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		log.Println("LogsProcesser.SaveDB()", err)
		return ethereumLogs, err
	}
	fromAddress = strings.ToLower(fromAddress)
	contractAddress = strings.ToLower(contractAddress)
	hash = strings.ToLower(hash)

	ethereumLogs.ChainId = chainId
	ethereumLogs.FromAddress = fromAddress
	ethereumLogs.ContractAddress = contractAddress
	ethereumLogs.Event = event
	ethereumLogs.BlockNumber = blockNumber
	ethereumLogs.LogIndex = logIndex
	ethereumLogs.Hash = hash
	ethereumLogs.Data = string(jsonStr)

	ethereumLogs, err = ethereumLogsDao.Create(ethereumLogs, nil)
	if err != nil {
		log.Println("LogsProcesser.SaveDB()", err)
		return ethereumLogs, err
	}

	return ethereumLogs, nil
}

func (processer *LogsProcesser) PubSub(chainId int, fromAddress string, contractAddress string, event string, blockNumber int64, logIndex int64, hash string, data map[string]interface{}) (*pubsub.PublishResult, error) {
	fromAddress = strings.ToLower(fromAddress)
	contractAddress = strings.ToLower(contractAddress)
	hash = strings.ToLower(hash)

	pubsubData := map[string]interface{}{}
	pubsubData["chain_id"] = chainId
	pubsubData["from_address"] = fromAddress
	pubsubData["contract_address"] = contractAddress
	pubsubData["event"] = event
	pubsubData["block_number"] = blockNumber
	pubsubData["log_index"] = logIndex
	pubsubData["hash"] = hash
	pubsubData["data"] = data
	jsonStr, err := json.Marshal(pubsubData)
	if err != nil {
		log.Println("LogsProcesser.PubSub()", err)
		return nil, err
	}
	log.Println(string(jsonStr))
	if processer.PubsubTopic != nil {
		res := processer.PubsubTopic.Publish(context.Background(), &pubsub.Message{Data: jsonStr})
		return res, nil
	}
	return nil, nil
}

func CreateEthereumTransaction(ethTransReq models.EthereumTransactions) (models.EthereumTransactions, error) {
	ethTrans := ethereumTransactionsDao.GetByHash(ethTransReq.Hash)
	if ethTrans.ID > 0 {
		return ethTrans, nil
	}
	ethTransReq.Status = -1
	ethTrans, err := ethereumTransactionsDao.Create(ethTransReq, nil)
	return ethTrans, err
}
