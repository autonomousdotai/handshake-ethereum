package main

import (
	"github.com/ninjadotorg/handshake-ethereum/controller"
	"github.com/ninjadotorg/handshake-ethereum/param"
	"github.com/ninjadotorg/handshake-ethereum/models"
	"log"
	"github.com/robfig/cron"
	"os"
	"github.com/gin-gonic/gin"
	"time"
	"net/http"
	"strconv"
	"github.com/urfave/cli"
	"io"
)

var (
	app *cli.App
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
	param.Initialize(os.Getenv("APP_CONF"))
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
		index.GET("/", func(context *gin.Context) {
			result := map[string]interface{}{
				"status":  1,
				"message": "Ethereum Service API",
			}
			context.JSON(http.StatusOK, result)
		})
		index.POST("/tx", func(context *gin.Context) {
			ethTrans := new(models.EthereumTransactions)
			err := context.Bind(&ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				context.JSON(http.StatusOK, result)
				return
			}

			_, err = controller.CreateEthereumTransaction(*ethTrans)
			if err != nil {
				result := map[string]interface{}{
					"status":  -1,
					"message": err.Error(),
				}
				context.JSON(http.StatusOK, result)
				return
			}

			result := map[string]interface{}{
				"status":  1,
				"message": "OK",
			}
			context.JSON(http.StatusOK, result)
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
		userId, _ := strconv.ParseInt(context.GetHeader("User-Id"), 10, 64)
		context.Set("UserId", userId)
		context.Next()
	}
}
