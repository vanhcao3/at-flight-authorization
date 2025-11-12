package service

import (
	"encoding/json"
	"strings"

	types "172.21.5.249/airtrans/at-flight-authorization/internal/types/authorization"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type EventRouting struct {
	ClientId string
}

func (s *Service) ProcessNatsMsg(m *nats.Msg) error {
	eventRouting, err := getEventRouting(m.Subject)
	if err != nil {
		return err
	}

	log.Debug().Msgf(
		"process message intended client: %s",
		eventRouting.ClientId,
	)

	msg := types.AuthorizationLog{}
	json.Unmarshal(m.Data, &msg)

	if err != nil {
		log.Error().Err(err).Msg("Unmarshal msg failed")
		return err
	}

	result := s.db.Create(msg)
	if result.Error != nil {
		log.Error().Err(err).Msg("Failed creating authorization log")
	}

	log.Debug().Msgf("Received data from client %s: %v", eventRouting.ClientId, msg)

	return nil
}

func getEventRouting(subject string) (*EventRouting, error) {
	elements := strings.Split(subject, ".")

	eventRouting := &EventRouting{
		ClientId: elements[1],
	}

	// if len(elements) > 4 {
	// 	eventRouting.ResourceName = elements[4]
	// }

	return eventRouting, nil
}
