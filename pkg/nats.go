package pkg

import "github.com/nats-io/nats.go"

type NatsPublisher struct {
	conn *nats.Conn
}

func NewNatsPublisher(natsUrl string) (*NatsPublisher, error) {
	conn, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, err
	}
	return &NatsPublisher{conn: conn}, nil
}

func (p *NatsPublisher) Publish(subject string, data []byte) error {
	return p.conn.Publish(subject, data)
}
