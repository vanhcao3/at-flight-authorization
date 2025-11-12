package config

type RedisConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
