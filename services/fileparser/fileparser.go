package fileparser

import (
	"encoding/json"
	"log"

	"github.com/coding-yogi/perftool/rmq"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

var exchangeName = "nameko-rpc"
var serviceName = "cv_parsing_service"
var rpcMethodName = "parse"

var (
	ch      *rmq.Channel
	replyTo string
	chMsgs  <-chan amqp.Delivery
	err     error
)

type Reply struct {
	ReplyTo string
	ChMsgs  <-chan amqp.Delivery
}

var sp = rmq.SetupParams{
	ServiceName:  serviceName,
	ExchangeName: exchangeName,
	ExchangeType: "topic",
	QueueName:    "rpc-" + serviceName,
	RoutingKey:   serviceName + "." + rpcMethodName,
}

type Kwargs struct {
}

type FileParserMessage struct {
	Kwargs Kwargs   `json:"kwargs"`
	Args   []string `json:"args"`
}

func Setup(amqpClient *rmq.Client) {
	ch, err = amqpClient.SetupChannel(sp)
	failOnError(err, "Unable to Setup Channel")
}

func (r *Reply) CreateQueue() {
	var err error
	r.ReplyTo, err = ch.CreateReplyQueue(sp.ExchangeName)
	failOnError(err, "Unable to create reply queue")

	r.ChMsgs, err = ch.ConsumeMessage(r.ReplyTo, "")
	failOnError(err, "Registering consumer failed")
}

func (r *Reply) Parse() {

	fileParserMessage := FileParserMessage{
		Kwargs: Kwargs{},
		Args:   []string{"c29tZSBmaWxlIHRleHQ=", "", "TXT"},
	}
	body, err := json.Marshal(fileParserMessage)
	failOnError(err, "Unable to create file parser message body")

	uuid := uuid.NewV4().String()
	log.Printf("Correlation ID:  %s", uuid)

	table := amqp.Table{}
	table["nameko.correlation_id"] = "111-222-333-444-555"

	err = ch.PublishMessageAsRPC(sp.ExchangeName, sp.RoutingKey, amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "UTF-8",
		CorrelationId:   uuid,
		DeliveryMode:    2,
		Priority:        0,
		ReplyTo:         r.ReplyTo,
		Body:            body,
		Headers:         table,
	})
	failOnError(err, "Publishing message failed")

	go func() {
		msg, _ := rmq.GetMessageForCorrelationID(r.ChMsgs, uuid)
		log.Printf("Msg response for correlation id %s -> %s", uuid, msg)
	}()

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
