package ledger

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"rounds/telemetry"
)

type Config struct {
	Ledger struct {
		Host string `validate:"required"`
	} `validate:"required"`
	DB struct {
		Path string `validate:"required"`
	} `validate:"required"`
	Opencensus telemetry.OpencensusConfig
	Logging    struct {
		Level    string `validate:"required"`
		Encoding string `validate:"required"`
	} `validate:"required"`
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
