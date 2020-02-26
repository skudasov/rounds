package node

import (
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log"
	"rounds/telemetry"
)

type Config struct {
	Node struct {
		Keyspath string `validate:"required"`
		Addr     string `validate:"required"`
		Peers    []Peer `validate:"required"`
		Rounds   struct {
			PaceMs  int `yaml:"paceMs"`
			Collect struct {
				MaxMessages int `json:"max_messages"`
				Duration    int `json:"duration"`
			} `validate:"required"`
			Exchange struct {
				MaxMessages int `json:"max_messages"`
				Duration    int `json:"duration"`
			} `validate:"required"`
		}
		Reconnect int    `json:"reconnect" validate:"required"`
		Transport string `json:"transport" validate:"required"`
	}
	Opencensus telemetry.OpencensusConfig
	Store      struct {
		Host string `validate:"required"`
	} `validate:"required"`
	Logging struct {
		Level string
	}
}

type Peer struct {
	Addr      string
	Port      string
	PubKeyDir string `json:"pubkeydir"`
}

type RoundConfigs struct {
	Rounds []RoundConfig
}

type RoundConfig struct {
	Name string
	Tick int
}

func MakeConfig() *Config {
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

func ValidateConfig(cfg interface{}) {
	if err := validator.New().Struct(cfg); err != nil {
		log.Fatal(err)
	}
}
