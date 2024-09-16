package main

import (
	"context"
	"log"

	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/synera-br/golang-cloud-collector/configs"

	_ "github.com/synera-br/golang-cloud-collector/docs/swagger"
	"github.com/synera-br/golang-cloud-collector/internal/core/repository"
	"github.com/synera-br/golang-cloud-collector/internal/core/service"
	handler "github.com/synera-br/golang-cloud-collector/internal/infra/handler/rest"
	"github.com/synera-br/golang-cloud-collector/pkg/cache"
	"github.com/synera-br/golang-cloud-collector/pkg/mq"
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
	http_server "github.com/synera-br/golang-cloud-collector/pkg/service_http/server"
)

// @title        cloud-collector-resources
// @version      1.0
// @description  This service collect the resources from cloud provider and convert to Backstage structure

// @contact.name   Rafael Tomelin
// @contact.url    https://local
// @contact.email  contato@synera.com.br

// @schemes   http
// @BasePath  /api
func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// STARTS OTEL
	ctx := context.Background()
	otl, err := otelpkg.NewOtel(ctx, cfg.FileConfig.ConfigPath, cfg.FileConfig.FileName, cfg.FileConfig.Extentsion)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, span := otl.Tracer.Start(ctx, "main")
	defer span.End()

	defer otl.TracerSdk.Shutdown(ctx)

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(otl.Parameters.AppName),
		newrelic.ConfigLicense(otl.Parameters.License),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		log.Fatalln(err)
	}
	// ENDS OTEL

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
	azureRepository, err := repository.NewAzureRepository(&cfg.Provider.Azure, otl)
	if err != nil {
		log.Fatalln(err)
	}

	azureService, err := service.NewAzureService(&azureRepository, &cc, otl)
	if err != nil {
		log.Fatalln(err)
	}

	handler.NewAzureHandlerHttp(azureService, otl, rest.RouterGroup, rest.ValidateToken)
	if err != nil {
		log.Fatalln("error is: ", err.Error())
	}

	// Backstage
	backstageService := service.NewBackstageService(azureService, amqp, cc, otl)
	handler.NewBackstageHandlerHttp(backstageService, otl, rest.RouterGroup, rest.ValidateToken, nrgin.Middleware(app))
	if err != nil {
		log.Fatalln("error is: ", err.Error())
	}

	rest.Run(rest.Route.Handler())

}
