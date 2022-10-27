package config

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type EnvType string

const (
	test EnvType = "test"
	prod EnvType = "prod"
	dev  EnvType = "dev"
)

type (
	Config struct {
		Environment EnvType `envconfig:"ENVIRONMENT" required:"true"`
		HTTP        HTTP
		Postgres    Postgres
		Report      Report
		Logger      Logger
	}

	HTTP struct {
		Host           string        `envconfig:"HTTP_HOST"             required:"true"`
		Port           string        `envconfig:"HTTP_PORT"             required:"true"`
		MaxHeaderBytes int           `envconfig:"HTTP_MAX_HEADER_BYTES"                 default:"1"`
		ReadTimeout    time.Duration `envconfig:"HTTP_READ_TIMEOUT"                     default:"10s"`
		WriteTimeout   time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"                    default:"10s"`
	}

	Postgres struct {
		Host        string `envconfig:"POSTGRES_HOST"          required:"true"`
		Port        string `envconfig:"POSTGRES_PORT"          required:"true"`
		DBName      string `envconfig:"POSTGRES_DBNAME"        required:"true"`
		User        string `envconfig:"POSTGRES_USER"          required:"true"`
		Password    string `envconfig:"POSTGRES_PASSWORD"      required:"true" json:"-"`
		SSLMode     string `envconfig:"POSTGRES_SSLMODE"                                default:"disable"`
		MaxPoolSize int    `envconfig:"POSTGRES_MAX_POOL_SIZE"                          default:"4"`
	}

	Report struct {
		Host string `envconfig:"REPORT_HOST" required:"true"`
		Port string `envconfig:"REPORT_PORT" required:"true"`
	}

	Logger struct {
		Level string `envconfig:"LOGGER_LEVEL" default:"info"`
	}
)

func (c *Config) IsDev() bool {
	return c.Environment == dev
}

func (c *Config) IsTest() bool {
	return c.Environment == test
}

func (c *Config) IsProd() bool {
	return c.Environment == prod
}

var (
	instance Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		if err := envconfig.Process("", &instance); err != nil {
			log.Fatal(err)
		}

		switch instance.Environment {
		case test, prod, dev:
		default:
			log.Fatal("config environment should be test, prod or dev")
		}
		if instance.IsDev() {
			configBytes, err := json.MarshalIndent(instance, "", " ")
			if err != nil {
				log.Fatal(fmt.Errorf("error marshaling indent config: %w", err))
			}
			fmt.Println("Configuration:", string(configBytes))
		}
	})

	return &instance
}
