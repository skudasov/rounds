package ledger

import (
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

type Config struct {
	Ledger struct {
		Host string
	}
	DB struct {
		Path string
	}
	Opencensus struct {
		Prometheus struct {
			Nodelabel string
			Port      string
		}
		Jaeger struct {
			Nodelabel string
			Port      string
		}
	}
	Logging struct {
		Level string
	}
}

func LoadConfig() *Config {
	c := flag.String("config", "node.yml", "path to node config file")
	flag.Parse()

	viper.SetConfigName(*c)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %s", err)
	}
	return cfg
}
