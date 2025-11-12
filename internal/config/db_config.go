package config

import "time"

type MongoConfig struct {
	//DB config
	DBHost    string `mapstructure:"host"`
	DBPort    int    `mapstructure:"port"`
	DBName    string `mapstructure:"name"`
	DBUser    string `mapstructure:"user"`
	DBPass    string `mapstructure:"pass"`
	DBReplica string `mapstructure:"replica"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Password        string        `mapstructure:"password"`
	User            string        `mapstructure:"user"`
	DBName          string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslmode"`
	Retention       time.Duration `mapstructure:"retention"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}
