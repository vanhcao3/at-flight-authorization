package config

type NatsConfig struct {
	BindAddress string `mapstructure:"bind_address"`
	NodeName    string `mapstructure:"node_name"`
}
