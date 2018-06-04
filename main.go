package main

import (
	"github.com/ninjadotorg/handshake-ethereum/controller"
	"github.com/ninjadotorg/handshake-ethereum/param"
	"log"
	"github.com/robfig/cron"
	"os"
)

func main() {
	param.Initialize(os.Getenv("APP_CONF"))
	controller, err := controller.NewConcotrller(param.Conf.Agrs)
	if err != nil {
		log.Print(err)
		return
	}
	var appCron = cron.New()
	appCron.AddFunc("0 * * * * *", func() {
		log.Println("job for scan ethereum logs")
		controller.Process()
	})
	appCron.Start()
	select {}
}
