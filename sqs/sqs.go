package sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SqsMessage interface {
	GetMsgAttributes() map[string]*sqs.MessageAttributeValue
	GetMsgBody() *string
	SetReceiveMsg(message *sqs.ReceiveMessageOutput) error
}

type sqsClient struct {
	svc *sqs.SQS
}

func NewSqsClient(sess *session.Session) *sqsClient {
	return &sqsClient{svc: sqs.New(sess)}
}

func (c *sqsClient) GetQueueURL(queue *string) (*sqs.GetQueueUrlOutput, error) {
	// Create an SQS service client
	result, err := c.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queue,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *sqsClient) sendMsg(queueURL *string, msg SqsMessage) error {
	// Create an SQS service client
	// snippet-start:[sqs.go.send_message.call]

	_, err := c.svc.SendMessage(&sqs.SendMessageInput{
		//DelaySeconds:      aws.Int64(10),
		MessageAttributes: msg.GetMsgAttributes(),
		MessageBody:       msg.GetMsgBody(),
		QueueUrl:          queueURL,
	})
	// snippet-end:[sqs.go.send_message.call]
	if err != nil {
		return err
	}

	return nil
}

func (c *sqsClient) SendMsg(queueName string, msg SqsMessage) error {
	result, err := c.GetQueueURL(&queueName)
	if err != nil {
		fmt.Printf("GetQueueURL error: %v", err)
		return err
	}

	err = c.sendMsg(result.QueueUrl, msg)
	if err != nil {
		fmt.Printf("SendMsg error: %v", err)
		return err
	}
	return nil
}

func (c *sqsClient) GetMessages(queueName string, output SqsMessage, timeout *int64) error {
	result, err := c.GetQueueURL(&queueName)
	if err != nil {
		fmt.Printf("GetQueueURL error: %v", err)
		return err
	}
	queueURL := result.QueueUrl

	// snippet-start:[sqs.go.receive_messages.call]
	msgResult, err := c.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   timeout,
	})
	// snippet-end:[sqs.go.receive_messages.call]
	if err != nil {
		return err
	}
	return output.SetReceiveMsg(msgResult)
}
