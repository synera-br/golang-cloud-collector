package http_server

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

type Certificate struct {
	Key string `mapstructure:"key"`
	Crt string `mapstructure:"crt"`
}

type RestAPIConfig struct {
	Port           string `mapstructure:"port"`
	SSLEnabled     bool   `mapstructure:"ssl_enabled"`
	Host           string `mapstructure:"host"`
	Version        string `mapstructure:"version"`
	Name           string `mapstructure:"name"`
	CertificateCrt string `mapstructure:"certificate_crt"`
	CertificateKey string `mapstructure:"certificate_key"`
	Token          string `mapstructure:"token"`
}

type RestAPI struct {
	Config *RestAPIConfig
	Route  *gin.Engine
	*gin.RouterGroup
}

func NewRestApi(pathConfigFile, nameFileConfig, nameFileExtention string) (*RestAPI, error) {

	viper.AddConfigPath(pathConfigFile)
	viper.SetConfigName(nameFileConfig)
	viper.SetConfigType(nameFileExtention)
	viper.AutomaticEnv()

	v, ok := viper.Get("webserver").(map[string]interface{})
	if !ok {
		log.Panicln("error to load http configurations")
	}

	rest := Parse(v)

	r, g := newRestAPI(rest)

	return &RestAPI{
		Config:      rest,
		Route:       r,
		RouterGroup: g,
	}, nil
}

func Parse(m map[string]interface{}) *RestAPIConfig {

	c := RestAPIConfig{
		Port:           "8080",
		SSLEnabled:     false,
		Host:           "0.0.0.0",
		Version:        "v1",
		Name:           "",
		CertificateCrt: "",
		CertificateKey: "",
		Token:          "",
	}

	if m["port"] != nil {
		if reflect.TypeOf(m["port"]).Kind() == reflect.String {
			c.Port = m["port"].(string)

		} else if reflect.TypeOf(m["port"]).Kind() == reflect.Int {
			c.Port = strconv.Itoa(m["port"].(int))
		}
	}

	if m["host"] != nil {
		c.Host = m["host"].(string)
	}

	if m["name"] != nil {
		c.Name = m["name"].(string)
	}

	if m["version"] != nil {
		c.Version = m["version"].(string)
	}

	if m["token"] != nil {
		c.Token = m["token"].(string)
	}

	if m["certificate_crt"] != nil {
		c.CertificateCrt = m["certificate_crt"].(string)
	}
	if m["certificate_key"] != nil {
		c.CertificateKey = m["certificate_key"].(string)
	}
	if m["ssl_enabled"] != nil {
		c.SSLEnabled = m["ssl_enabled"].(bool)
	}

	return &c
}

var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	// SwaggerTemplate:  docTemplate,
	LeftDelim:  "{{",
	RightDelim: "}}",
}

func newRestAPI(config *RestAPIConfig) (*gin.Engine, *gin.RouterGroup) {

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.UseH2C = true

	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1", "192.168.1.2", "10.0.0.0/8"})

	routerGroupPath := fmt.Sprintf("/%s", config.Name)
	routerPath = router.Group(routerGroupPath)

	router.GET("/metrics", prometheusHandler())

	// Set swagger

	routerPath.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.DefaultModelsExpandDepth(-1)))
	routerPath.GET("/docs/swagger", func(c *gin.Context) {
		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})
	routerPath.GET("/docs", func(c *gin.Context) {

		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})
	routerPath.GET("/", func(c *gin.Context) {
		c.Redirect(301, fmt.Sprintf("%s/docs/swagger/index.html", routerGroupPath))
	})

	router.Use(setHeader)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(cors.New(corsConfig))

	return router, routerPath

}
