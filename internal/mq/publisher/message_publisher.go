package publisher

import (
	"context"

	"github.com/rs/zerolog/log"

	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type MessagePublisher struct {
	Publisher    *rabbitmq.Publisher
	ExchangeName string
}

func NewPublisher(rbConn *rabbitmq.Conn, exchangeName string) *MessagePublisher {
	publisher, err := rabbitmq.NewPublisher(
		rbConn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(exchangeName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeDurable,
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Can not start publisher!")
	}

	return &MessagePublisher{
		Publisher:    publisher,
		ExchangeName: exchangeName,
	}
}

func (mp *MessagePublisher) Publish(ctx context.Context, data []byte, key []string) {
	err := mp.Publisher.PublishWithContext(
		ctx,
		data,
		key,
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(mp.ExchangeName),
	)
	mp.Publisher.NotifyPublish(func(c rabbitmq.Confirmation) {
		log.Debug().Msgf("message confirmed from rabbitmq tag: %v, ack: %v", c.DeliveryTag, c.Ack)
	})
	mp.Publisher.NotifyReturn(func(r rabbitmq.Return) {
		log.Debug().Msgf("message returned from rabbitmq: %s", string(r.Body))
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to publish event message with keys %v", key)
	}
}
