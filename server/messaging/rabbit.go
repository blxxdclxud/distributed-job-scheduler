package messaging

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type RabbitMQPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	exchangeType string
	logger       *slog.Logger
}

func NewRabbitMQPublisher(
	conn *amqp.Connection,
	exchangeName string,
	exchangeType string,
) (*RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		return nil, err
	}

	return &RabbitMQPublisher{
		conn:         conn,
		channel:      channel,
		exchangeName: exchangeName,
		exchangeType: exchangeType,
	}, nil
}

func (p *RabbitMQPublisher) PublishJSON(
	ctx context.Context,
	routingKey string,
	message interface{},
) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
