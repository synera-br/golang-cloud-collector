package main

import (
	"log"

	"github.com/synera-br/golang-cloud-collector/configs"
	_ "github.com/synera-br/golang-cloud-collector/docs/swagger"
	"github.com/synera-br/golang-cloud-collector/internal/core/repository"
	"github.com/synera-br/golang-cloud-collector/internal/core/service"
	handler "github.com/synera-br/golang-cloud-collector/internal/infra/handler/rest"
	"github.com/synera-br/golang-cloud-collector/pkg/cache"
	"github.com/synera-br/golang-cloud-collector/pkg/mq"
	http_server "github.com/synera-br/golang-cloud-collector/pkg/service_http/server"
)

// @title        cloud-collector-resources
// @version      1.0
// @description  This service collect the resources from cloud provider and convert to Backstage structure

// @contact.name   Rafael Tomelin
// @contact.url    https://local
// @contact.email  rafael.tomelin@gmail.com

// @schemes   http
// @BasePath  /api
func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// Inicia os serviços
	cc, err := cache.NewCacheConnection(cfg.FileConfig.ConfigPath, cfg.FileConfig.FileName, cfg.FileConfig.Extentsion)
	if err != nil {
		log.Fatalln(err)
	}

	rest, err := http_server.NewRestApi(cfg.FileConfig.ConfigPath, cfg.FileConfig.FileName, cfg.FileConfig.Extentsion)
	if err != nil {
		log.Fatalln(err)
	}

	amqp, err := mq.NewMQConnection(cfg.FileConfig.ConfigPath, cfg.FileConfig.FileName, cfg.FileConfig.Extentsion)
	if err != nil {
		log.Fatalln(err)
	}

	// Azure resources
	azureRepository, err := repository.NewAzureRepository(&cfg.Provider.Azure)
	if err != nil {
		log.Fatalln(err)
	}

	azureService, err := service.NewAzureService(&azureRepository, &cc)
	if err != nil {
		log.Fatalln(err)
	}

	handler.NewAzureHandlerHttp(azureService, rest.RouterGroup, rest.ValidateToken)
	if err != nil {
		log.Fatalln("error is: ", err.Error())
	}

	// Backstage
	backstageService := service.NewBackstageService(azureService, amqp, cc)
	handler.NewBackstageHandlerHttp(backstageService, rest.RouterGroup, rest.ValidateToken)
	if err != nil {
		log.Fatalln("error is: ", err.Error())
	}
	rest.Run(rest.Route.Handler())

}
