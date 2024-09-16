package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

type AMQPServiceInterface interface {
	Consumer(ctx context.Context, data DataAMQP, msgChannel chan<- amqp.Delivery) error
	Publish(ctx context.Context, data DataAMQP) error
	Close() error
}

func (a *MQConfig) Consumer(ctx context.Context, data DataAMQP, msgChannel chan<- amqp.Delivery) error {

	msgs, err := a.Channel.Consume(
		data.Queue, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {

			msgChannel <- msg
		}
		close(msgChannel)
	}()

	return err

}

// ProducerExhange
// Send a message to amqp server
// context
// exchange to connect
// queue to send a message
// message to send
// route key to apply in message
func (a *MQConfig) Publish(ctx context.Context, data DataAMQP) error {

	tracer := otel.Tracer("mq.Publish")
	_, span := tracer.Start(ctx, "mq.Publish")
	defer span.End()

	headers := amqp.Table{
		"trace-id": span.SpanContext().TraceID().String(),
		"span-id":  span.SpanContext().SpanID().String(),
	}

	content := "text/plain"
	if data.ContentType != "" {
		content = data.ContentType
	}

	exchange := ""
	if data.Exchange != "" {
		exchange = data.Exchange
	}
	err := a.Channel.PublishWithContext(ctx,
		exchange,      // exchange
		data.RouteKey, // routing key
		false,         // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  content,
			Body:         data.Body,
			Headers:      headers,
		})

	return err
}

func (a *MQConfig) Close() error {
	return a.Channel.Close()
}

func (a *MQConfig) SetupExchangeQueueAndBind() error {

	for _, r := range a.Rules.Exchanges {
		// Declare a Exchange
		err := a.Channel.ExchangeDeclare(
			r.Name,
			r.Type,
			r.Durable,
			r.AutoDelete, // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}

	}

	for _, r := range a.Rules.Queues {
		args := amqp.Table{}
		if r.DeadLetterExchange != "" {
			args["x-dead-letter-exchange"] = r.DeadLetterExchange
			if r.DeadLetterRoutingKey != "" {
				args["x-dead-letter-routing-key"] = r.DeadLetterRoutingKey
			}
		}

		// Declare a queue
		_, err := a.Channel.QueueDeclare(
			r.Name,       // name
			r.Durable,    // durable
			r.AutoDelete, // delete when unused
			r.Exclusive,  // exclusive
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}
	}

	for _, r := range a.Rules.Bindings {
		// Declare a bindig
		err := a.Channel.QueueBind(
			r.Queue,      // queue name
			r.RoutingKey, // routing key
			r.Exchange,   // exchange name
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
	}

	return nil
}
