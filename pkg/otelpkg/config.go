package otelpkg

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type otelConfig struct {
	Endpoint string            `mapstructure:"endpoint"`
	Headers  map[string]string `mapstructure:"headers"`
	Protocol string            `mapstructure:"protocol"`
	Insecure bool              `mapstructure:"insecure"`
	Encoding string            `mapstructure:"encoding"`
	Name     string            `mapstructure:"name"`
	Provider string            `mapstructure:"provider"`
}

func parseConfig(pathConfigFile, nameFileConfig, nameFileExtention string) (*otelConfig, error) {

	viper.AddConfigPath(pathConfigFile)
	viper.SetConfigName(nameFileConfig)
	viper.SetConfigType(nameFileExtention)
	viper.AutomaticEnv()

	v, ok := viper.Get("otel").(map[string]interface{})
	if !ok {
		log.Panicln("error to load http configurations")
	}

	rest := parse(v)
	if rest == nil {
		return nil, errors.New("endpoint otel did not definied")
	}
	return rest, nil

}

func parse(m map[string]interface{}) *otelConfig {

	c := otelConfig{
		Endpoint: "",
		Headers:  make(map[string]string),
		Protocol: "https",
		Insecure: true,
		Encoding: "json",
		Name:     "myapp",
		Provider: "console",
	}

	if m["headers"] != nil {
		for k, v := range m["headers"].(map[string]interface{}) {
			c.Headers[k] = fmt.Sprintf("%v", v)
		}
	}

	if m["protocol"] != nil {
		c.Protocol = m["protocol"].(string)
	}

	if m["insecure"] != nil {
		c.Insecure = m["insecure"].(bool)
	}

	if m["encoding"] != nil {
		c.Encoding = m["encoding"].(string)
	}

	if m["name"] != nil {
		c.Name = m["name"].(string)
	}

	if m["provider"] != nil {
		c.Provider = m["provider"].(string)
	} else {
		c.Provider = "console"
	}

	if c.Provider != "console" {
		if m["endpoint"] == nil || m["endpoint"] == "" {
			return nil
		} else {
			c.Endpoint = m["endpoint"].(string)
		}
	}

	return &c
}
