package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

type DatabaseConfig struct {
	Driver         string `mapstructure:"driver"`
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Name           string `mapstructure:"name"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers   []string `mapstructure:"brokers"`
	Topic     string   `mapstructure:"topic"`
	Acks      int16    `mapstructure:"acks"`
	LingerMs  int64    `mapstructure:"linger_ms"`
	BatchSize int32    `mapstructure:"batch_size"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./")

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8080")

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	//v.AutomaticEnv()
	err := v.BindEnv("database.password", "APP_DATABASE_PASSWORD")
	if err != nil {
		return nil, err
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
func (d *DatabaseConfig) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name)
}
