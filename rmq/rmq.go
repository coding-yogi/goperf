package rmq

import (
	"fmt"
	"log"
	"math/rand"

	"time"

	"github.com/streadway/amqp"
)

type IClient interface {
	NewConnection() error
	CloseConnection() error
	CreateChannel() (*amqp.Channel, error)
}

type IChannel interface {
	CreateExchange(exName, exType string) error
	CreateQueue(queueName string) (amqp.Queue, error)
	BindQueueToExchange(queueName, routingKey, exName string) error
	PublishMessageAsRPC(exName, routingKey string, props amqp.Publishing) error
	GetRandomQueue() error
}

type Client struct {
	Host     string
	Port     int32
	Username string
	Password string
	conn     *amqp.Connection
}

type SetupParams struct {
	ServiceName  string
	ExchangeName string
	ExchangeType string
	QueueName    string
	RoutingKey   string
}

type Channel struct {
	*amqp.Channel
}

// NewConnection ...
func (c *Client) NewConnection() error {
	var err error
	c.conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", c.Username, c.Password, c.Host, c.Port))
	return err
}

// Setup
func (c *Client) SetupChannel(params SetupParams) (*Channel, error) {
	ch, err := c.createChannel()
	if err != nil {
		return nil, err
	}

	err = ch.createExchange(params.ExchangeName, params.ExchangeType)
	if err != nil {
		return nil, err
	}

	_, err = ch.createQueue(params.QueueName)
	if err != nil {
		return nil, err
	}

	err = ch.bindQueueToExchange(params.QueueName, params.RoutingKey, params.ExchangeName)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// CloseConnection ...
func (c *Client) CloseConnection() error {
	return c.conn.Close()
}

// CreateChannel ...
func (c *Client) createChannel() (*Channel, error) {
	log.Println("Creating Channel")
	ch, err := c.conn.Channel()
	return &Channel{ch}, err
}

// CreateExchange ...
func (ch *Channel) createExchange(exName, exType string) error {
	log.Printf("Creating Exchange by name %s of type %s", exName, exType)
	return ch.ExchangeDeclare(exName, exType, true, false, false, false, nil)
}

// CreateQueue  ...
func (ch *Channel) createQueue(queueName string) (amqp.Queue, error) {
	log.Printf("Creating Queue by name %s ", queueName)
	return ch.QueueDeclare(queueName, true, false, false, false, nil)
}

// BindQueueToExchange ...
func (ch *Channel) bindQueueToExchange(queueName, routingKey, exName string) error {
	log.Printf("Binding Queue %s to Exchange %s ", queueName, exName)
	return ch.QueueBind(queueName, routingKey, exName, false, nil)
}

// PublishMessageAsRPC ...
func (ch *Channel) PublishMessageAsRPC(exName, routingKey string, msg amqp.Publishing) error {
	log.Printf("Publish message to Exchange %s with Routing Key %s ", exName, routingKey)
	return ch.Publish(exName, routingKey, false, false, msg)
}

// GetRandomQueue ...
func (ch *Channel) GetRandomQueue() (string, error) {
	queue, err := ch.QueueDeclare("", false, false, true, false, nil)
	log.Printf("Created Random Queue by name %s ", queue.Name)
	return queue.Name, err
}

// Consume ...
func (ch *Channel) ConsumeMessage(queueName, consumer string) (<-chan amqp.Delivery, error) {
	return ch.Consume(queueName, consumer, true, false, false, false, nil)
}

// CreateReplyQueue
func (ch *Channel) CreateReplyQueue(exName string) (string, error) {
	replyQueue, err := ch.GetRandomQueue()
	if err != nil {
		return "", err
	}

	err = ch.bindQueueToExchange(replyQueue, replyQueue, exName)
	if err != nil {
		return "", err
	}

	return replyQueue, nil
}

// GetMessageForCorrelationID ...
func GetMessageForCorrelationID(msgs <-chan amqp.Delivery, corrID string) ([]byte, error) {
	for msg := range msgs {
		if msg.CorrelationId == corrID {
			return msg.Body, nil
		}

		log.Printf("Waiting for message with corr ID %s but received msg with corr ID %s", corrID, msg.CorrelationId)
	}

	return nil, fmt.Errorf("No message found for correlation id %s", corrID)
}

func randomQueueName(n int) string {

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	src := rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
