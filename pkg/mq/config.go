package mq

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

type DataAMQP struct {
	ContentType string
	Body        []byte
	Exchange    string
	Queue       string
	RouteKey    string
}
type Rules struct {
	Exchanges []Exchange `json:"exchanges" mapstructure:"exchanges"`
	Queues    []Queue    `json:"queues" mapstructure:"queues"`
	Bindings  []Binding  `json:"bindings" mapstructure:"bindings"`
}

type Exchange struct {
	Name       string `json:"name" mapstructure:"name"`
	Type       string `json:"type" mapstructure:"type"`
	Durable    bool   `json:"durable" mapstructure:"durable"`
	AutoDelete bool   `json:"auto_delete" mapstructure:"auto_delete"`
}

type Queue struct {
	Name                 string `json:"name" mapstructure:"name"`
	Durable              bool   `json:"durable" mapstructure:"durable"`
	Exclusive            bool   `json:"exclusive" mapstructure:"exclusive"`
	AutoDelete           bool   `json:"auto_delete" mapstructure:"auto_delete"`
	DeadLetterExchange   string `json:"dead_letter_exchange" mapstructure:"dead_letter_exchange"`
	DeadLetterRoutingKey string `json:"dead_letter_routing_key" mapstructure:"dead_letter_routing_key"`
}

type Binding struct {
	Queue      string `json:"queue" mapstructure:"queue"`
	Exchange   string `json:"exchange" mapstructure:"exchange"`
	RoutingKey string `json:"routing_key" mapstructure:"routing_key"`
}

type MQConfig struct {
	Host       string `json:"host" mapstructure:"host"`
	User       string `json:"user" mapstructure:"user"`
	Password   string `json:"password" mapstructure:"password"`
	SSLEnabled bool   `json:"ssl_enabled" mapstructure:"ssl_enabled"`
	Port       string `json:"port" mapstructure:"port"`
	Rules      Rules  `json:"rules" mapstructure:"rules"`
	VHost      string `json:"vhost" mapstructure:"vhost"`
	Ttl        int    `json:"ttl" mapstructure:"ttl"`
	Channel    *amqp.Channel
}

func NewMQConnection(pathConfigFile, nameFileConfig, nameFileExtention string) (AMQPServiceInterface, error) {

	cfg := Parse(pathConfigFile, nameFileConfig, nameFileExtention)
	protocol := "amqp"
	if cfg.SSLEnabled {
		protocol = "amqps"

	}
	conn, err := amqp.Dial(fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.VHost))
	if err != nil {
		return nil, err
	}

	cfg.Channel, err = conn.Channel()
	if err != nil {
		return nil, err
	}

	err = cfg.SetupExchangeQueueAndBind()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func Parse(pathConfigFile, nameFileConfig, nameFileExtention string) *MQConfig {

	viper.AddConfigPath(pathConfigFile)
	viper.SetConfigName(nameFileConfig)
	viper.SetConfigType(nameFileExtention)
	viper.AutomaticEnv()

	m, ok := viper.Get("amqp").(map[string]interface{})
	if !ok {
		log.Panicln("error to load http configurations")
	}

	c := MQConfig{
		Port:       "5672",
		SSLEnabled: false,
		Host:       "localhost",
		User:       "",
		Password:   "",
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

	if m["rules"] != nil {

		b, err := json.Marshal(m["rules"])
		if err != nil {
			return nil
		}
		var rules Rules
		err = json.Unmarshal(b, &rules)
		if err != nil {
			return nil
		}

		c.Rules = rules
	}

	if m["vhost"] != nil {
		c.VHost = m["vhost"].(string)
	}

	if m["ttl"] != nil {
		c.Ttl = m["ttl"].(int)
	}

	c.SSLEnabled = m["ssl_enabled"].(bool)

	return &c
}
