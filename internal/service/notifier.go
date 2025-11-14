package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

type NotificationEvent string

const (
	EventFlightAuthorizationProposalCreated NotificationEvent = "flight_authorization_proposal.created"
	EventFlightAuthorizationApprovalCreated NotificationEvent = "flight_authorization_approval.created"
	EventFlightNotificationCreated          NotificationEvent = "flight_notification.created"
)

type eventMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Notifier struct {
	conn *nats.Conn
}

func NewNotifier(conn *nats.Conn) *Notifier {
	return &Notifier{
		conn: conn,
	}
}

func (n *Notifier) subject(event NotificationEvent) string {
	return fmt.Sprintf("websocket.%s", event)
}

func (n *Notifier) Subscribe(event NotificationEvent) (<-chan []byte, func(), error) {
	noop := func() {}
	if n == nil || n.conn == nil {
		return nil, noop, errors.New("notifier connection not initialized")
	}
	messages := make(chan []byte, 16)
	stop := make(chan struct{})
	sub, err := n.conn.Subscribe(n.subject(event), func(msg *nats.Msg) {
		data := make([]byte, len(msg.Data))
		copy(data, msg.Data)
		select {
		case messages <- data:
		case <-stop:
		}
	})
	if err != nil {
		close(messages)
		close(stop)
		return nil, noop, err
	}
	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			close(stop)
			_ = sub.Unsubscribe()
			close(messages)
		})
	}
	return messages, unsubscribe, nil
}

func (n *Notifier) Publish(event NotificationEvent, payload interface{}) error {
	if n == nil || n.conn == nil {
		return errors.New("notifier connection not initialized")
	}
	msg := eventMessage{
		Type: string(event),
		Data: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return n.conn.Publish(n.subject(event), data)
}
