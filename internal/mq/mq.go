package mq

import (
	"fmt"

	"172.21.5.249/airtrans/at-flight-authorization/internal/config"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

func NewConnection(cfg config.RabbitMQConfig) *rabbitmq.Conn {
	url := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s",
		cfg.Schema,
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	)
	conn, err := rabbitmq.NewConn(
		url,
		rabbitmq.WithConnectionOptionsLogging,
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{Properties: amqp091.Table{"connection_name": config.GetModuleName()}}),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Can not connect to rabbitmq")
	}
	return conn
}
