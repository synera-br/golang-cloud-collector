package cache

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type CacheConfig struct {
	Host       string `json:"" mapstructure:"host"`
	User       string `json:"user" mapstructure:"user"`
	Password   string `json:"password" mapstructure:"password"`
	SSLEnabled bool   `json:"ssl_enabled" mapstructure:"ssl_enabled"`
	Port       string `json:"port" mapstructure:"port"`
	Database   string `json:"database" mapstructure:"database"`
	Prefix     string `json:"prefix" mapstructure:"prefix"`
	Ttl        int    `json:"ttl" mapstructure:"ttl"`
	Client     *redis.Client
}

func NewCacheConnection(pathConfigFile, nameFileConfig, nameFileExtention string) (CacheInterface, error) {

	cfg := Parse(pathConfigFile, nameFileConfig, nameFileExtention)

	opt, err := redis.ParseURL(fmt.Sprintf("redis://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database))
	if err != nil {
		panic(err)
	}

	cfg.Client = redis.NewClient(opt)

	// Enable tracing instrumentation.
	// if err := redisotel.InstrumentTracing(cfg.Client, ); err != nil {
	// 	panic(err)
	// }

	// Enable metrics instrumentation.
	// if err := redisotel.InstrumentMetrics(cfg.Client); err != nil {
	// 	panic(err)
	// }

	return cfg, nil
}

func Parse(pathConfigFile, nameFileConfig, nameFileExtention string) *CacheConfig {

	viper.AddConfigPath(pathConfigFile)
	viper.SetConfigName(nameFileConfig)
	viper.SetConfigType(nameFileExtention)
	viper.AutomaticEnv()

	m, ok := viper.Get("cache").(map[string]interface{})
	if !ok {
		log.Panicln("error to load http configurations")
	}

	c := CacheConfig{
		Port:       "6379",
		SSLEnabled: false,
		Host:       "localhost",
		User:       "",
		Password:   "",
		Database:   "0",
		Prefix:     "app",
		Ttl:        60,
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

	if m["user"] != nil {
		c.User = m["user"].(string)
	}

	if m["password"] != nil {
		c.Password = m["password"].(string)
	}

	if m["database"] != nil {
		c.Database = m["database"].(string)
	}

	if m["prefix"] != nil {
		c.Prefix = m["prefix"].(string)
	}

	if m["ttl"] != nil {
		c.Ttl = m["ttl"].(int)
	}

	c.SSLEnabled = m["ssl_enabled"].(bool)

	return &c
}
