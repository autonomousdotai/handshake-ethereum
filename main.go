package main

import (
	"context"
	"crypto/ecdsa"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/ninjadotorg/handshake-ethereum/controller"
	"github.com/ninjadotorg/handshake-ethereum/models"
	"github.com/ninjadotorg/handshake-ethereum/param"
	"github.com/robfig/cron"
	"github.com/urfave/cli"
)

var (
	app          *cli.App
	etherClients = map[string]*ethclient.Client{}
)

func init() {
	// Initialise a CLI app
	app = cli.NewApp()
	app.Name = "ninja ethereum"
	app.Usage = "ninja ethereum"
	app.Author = "hieuqautonomous"
	app.Email = "hieu.q@autonomous.nyc"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Value: "",
			Usage: "Path to a configuration file",
		},
	}
}

func main() {
	app.Commands = []cli.Command{
		{
			Name:  "worker",
			Usage: "launch worker",
			Action: func(c *cli.Context) error {
				return workerApp()
			},
		},
		{
			Name:  "service",
			Usage: "launch service",
			Action: func(c *cli.Context) error {
				return serviceApp()
			},
		},
	}
	// Run the CLI app
	if err := app.Run(os.Args); err != nil {
		log.Println("error", err)
	}
	select {}
}

func workerApp() error {

	param.Initialize(os.Getenv("APP_CONF"))
	controller, err := controller.NewConcotrller(param.Conf.Agrs)
	if err != nil {
		log.Print(err)
		return err
	}
	var appCron = cron.New()
	appCron.AddFunc("*/16 * * * * *", func() {
		log.Println("job for scan ethereum logs every 16s")
		controller.Process()
	})
	appCron.Start()

	return nil
}

func serviceApp() error {
	err := param.Initialize(os.Getenv("APP_CONF"))
	if err != nil {
		panic(err)
	}

	for k, network := range param.Conf.Networks {
		etherClient, err := ethclient.Dial(network.NetworkURL)
		if err != nil {
			panic(err)
		}
		etherClients[k] = etherClient
	}

	// Logger
	logFile, err := os.OpenFile("logs/autonomous_service.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(gin.DefaultWriter) // You may need this
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	router := gin.Default()
	router.Use(Logger())
	router.Use(AuthorizeMiddleware())
	index := router.Group("/")
	{
		index.GET("/", func(c *gin.Context) {
			result := map[string]interface{}{
				"status":  1,
				"message": "Ethereum Service API",
			}
			c.JSON(http.StatusOK, result)
		})
		index.POST("/tx", func(c *gin.Context) {

			userID, ok := c.Get("UserID")
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			if userID.(int64) <= 0 {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			ethTrans := new(models.EthereumTransactions)
			err := c.Bind(&ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			ethTrans.UserID = userID.(int64)
			_, err = controller.CreateEthereumTransaction(*ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			result := map[string]interface{}{
				"status":  1,
				"message": "OK",
			}
			c.JSON(http.StatusOK, result)
		})
		index.POST("/transfer", func(c *gin.Context) {
			userID, ok := c.Get("UserID")
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			if userID.(int64) <= 0 {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			networkIDStr := c.Query("network_id")
			if networkIDStr == "" {
				networkIDStr = "rinkeby"
			}

			etherClient, ok := etherClients[networkIDStr]
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "network_id is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			privateKeyStr := c.Query("private_key")
			if privateKeyStr == "" {
				result := map[string]interface{}{
					"status":  -1,
					"message": "private_key is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			toAddressStr := c.Query("to_address")
			if toAddressStr == "" {
				result := map[string]interface{}{
					"status":  -1,
					"message": "to_address is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			valueFloat, err := strconv.ParseFloat(c.Query("value"), 64)

			privateKey, err := crypto.HexToECDSA(privateKeyStr)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "error casting public key to ECDSA",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

			nonce, err := etherClient.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			value := big.NewInt(int64(valueFloat * 1e18))
			gasLimit := uint64(100000) // in units
			gasPrice, err := etherClient.SuggestGasPrice(context.Background())
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			toAddress := common.HexToAddress(toAddressStr)

			tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
			signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			err = etherClient.SendTransaction(context.Background(), signedTx)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			ethTrans := models.EthereumTransactions{}
			ethTrans.Hash = signedTx.Hash().Hex()
			ethTrans.FromAddress = fromAddress.Hex()
			ethTrans.ToAddress = toAddressStr
			ethTrans.RefType = "user_transfer"
			ethTrans.RefID = userID.(int64)
			ethTrans.UserID = userID.(int64)
			ethTrans.Network = networkIDStr

			_, err = controller.CreateEthereumTransaction(ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			result := map[string]interface{}{
				"status": 1,
				"data": map[string]interface{}{
					"from_address": fromAddress.Hex(),
					"to_address":   toAddressStr,
					"hash":         signedTx.Hash().Hex(),
					"value":        valueFloat,
				},
			}
			c.JSON(http.StatusOK, result)
			return
		})

		index.POST("/free-ether", func(c *gin.Context) {
			userID, ok := c.Get("UserID")
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			if userID.(int64) <= 0 {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			networkIDStr := c.Query("network_id")
			if networkIDStr == "" {
				networkIDStr = "rinkeby"
			}
			network, ok := param.Conf.Networks[networkIDStr]
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "network_id is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			etherClient, ok := etherClients[networkIDStr]
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "network_id is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			toAddressStr := c.Query("to_address")
			if toAddressStr == "" {
				result := map[string]interface{}{
					"status":  -1,
					"message": "to_address is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			valueFloat, err := strconv.ParseFloat(c.Query("value"), 64)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			privateKeyStr := network.PrivateKey

			privateKey, err := crypto.HexToECDSA(privateKeyStr)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "error casting public key to ECDSA",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

			nonce, err := etherClient.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			value := big.NewInt(int64(valueFloat * 1e18))
			gasLimit := uint64(100000) // in units
			gasPrice, err := etherClient.SuggestGasPrice(context.Background())
			gasPrice = big.NewInt(gasPrice.Int64() + int64(5*1e09))
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			toAddress := common.HexToAddress(toAddressStr)

			tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
			signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			err = etherClient.SendTransaction(context.Background(), signedTx)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			ethTrans := models.EthereumTransactions{}
			ethTrans.Hash = signedTx.Hash().Hex()
			ethTrans.FromAddress = fromAddress.Hex()
			ethTrans.ToAddress = toAddressStr
			ethTrans.RefType = "user_free_ether"
			ethTrans.RefID = userID.(int64)
			ethTrans.UserID = userID.(int64)
			ethTrans.Network = networkIDStr

			_, err = controller.CreateEthereumTransaction(ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			result := map[string]interface{}{
				"status": 1,
				"data": map[string]interface{}{
					"from_address": fromAddress.Hex(),
					"to_address":   toAddressStr,
					"hash":         signedTx.Hash().Hex(),
					"value":        valueFloat,
				},
			}
			c.JSON(http.StatusOK, result)
			return
		})

		index.POST("/free-token", func(c *gin.Context) {
			userID, ok := c.Get("UserID")
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			if userID.(int64) <= 0 {
				result := map[string]interface{}{
					"status":  -1,
					"message": "user is not logged in",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			networkIDStr := c.Query("network_id")
			if networkIDStr == "" {
				networkIDStr = "rinkeby"
			}
			network, ok := param.Conf.Networks[networkIDStr]
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "network_id is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			etherClient, ok := etherClients[networkIDStr]
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "network_id is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			toAddressStr := c.Query("to_address")
			if toAddressStr == "" {
				result := map[string]interface{}{
					"status":  -1,
					"message": "to_address is invalid",
				}
				c.JSON(http.StatusOK, result)
				return
			}
			amountFloat, err := strconv.ParseFloat(c.Query("amount"), 64)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			privateKeyStr := network.PrivateKey
			tokenAddressStr := network.TokenAddress

			privateKey, err := crypto.HexToECDSA(privateKeyStr)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}
			publicKey := privateKey.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				result := map[string]interface{}{
					"status":  -1,
					"message": "error casting public key to ECDSA",
				}
				c.JSON(http.StatusOK, result)
				return
			}

			fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

			nonce, err := etherClient.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			value := big.NewInt(int64(0))
			gasLimit := uint64(100000) // in units
			gasPrice, err := etherClient.SuggestGasPrice(context.Background())
			gasPrice = big.NewInt(gasPrice.Int64() + int64(5*1e09))
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			toAddress := common.HexToAddress(toAddressStr)
			tokenAddress := common.HexToAddress(tokenAddressStr)

			transferFnSignature := []byte("transfer(address,uint256)")
			hash := sha3.NewKeccak256()
			hash.Write(transferFnSignature)
			methodID := hash.Sum(nil)[:4]

			paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

			amount := big.NewInt(int64(1e18 * amountFloat))
			paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

			var data []byte
			data = append(data, methodID...)
			data = append(data, paddedAddress...)
			data = append(data, paddedAmount...)

			// gasLimit, err = etherClient.EstimateGas(context.Background(), ethereum.CallMsg{
			// 	To:   &toAddress,
			// 	Data: data,
			// })

			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
			signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			err = etherClient.SendTransaction(context.Background(), signedTx)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			ethTrans := models.EthereumTransactions{}
			ethTrans.Hash = signedTx.Hash().Hex()
			ethTrans.FromAddress = fromAddress.Hex()
			ethTrans.ToAddress = toAddressStr
			ethTrans.RefType = "user_free_token"
			ethTrans.RefID = userID.(int64)
			ethTrans.UserID = userID.(int64)
			ethTrans.Network = networkIDStr

			_, err = controller.CreateEthereumTransaction(ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				c.JSON(http.StatusOK, result)
				return
			}

			result := map[string]interface{}{
				"status": 1,
				"data": map[string]interface{}{
					"token_address": tokenAddressStr,
					"from_address":  fromAddress.Hex(),
					"to_address":    toAddressStr,
					"hash":          signedTx.Hash().Hex(),
					"amount":        amountFloat,
				},
			}
			c.JSON(http.StatusOK, result)
			return
		})
	}

	router.Run(":8080")

	return nil
}

func Logger() gin.HandlerFunc {
	return func(context *gin.Context) {
		t := time.Now()
		context.Next()
		status := context.Writer.Status()
		latency := time.Since(t)
		log.Print("Request: " + context.Request.URL.String() + " | " + context.Request.Method + " - Status: " + strconv.Itoa(status) + " - " +
			latency.String())
	}
}

func AuthorizeMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		userID, _ := strconv.ParseInt(context.GetHeader("Uid"), 10, 64)
		if userID <= 0 {
			context.JSON(http.StatusOK, gin.H{"status": 0, "message": "User is not authorized"})
			context.Abort()
			return
		}
		context.Set("UserID", userID)
		context.Next()
	}
}
