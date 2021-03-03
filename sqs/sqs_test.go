package sqs

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws-opt-go/utils/sess"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var s *session.Session

func init() {
	s, _ = sess.NewSessionWithRegion("us-east-1")
}

type SqsMessagePicTest struct {
	UUID      string `json:"uuid"`
	OriobjURL string `json:"ori_obj_url"`
}

func (t *SqsMessagePicTest) GetMsgAttributes() map[string]*sqs.MessageAttributeValue {
	return map[string]*sqs.MessageAttributeValue{
		"Title": &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String("The Whistler"),
		},
		"Author": &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String("John Grisham"),
		},
		"WeeksOn": &sqs.MessageAttributeValue{
			DataType:    aws.String("Number"),
			StringValue: aws.String("6"),
		},
	}
}
func (t *SqsMessagePicTest) GetMsgBody() *string {
	return aws.String("This is a unit test.")
}

func (t *SqsMessagePicTest) SetReceiveMsg(message *sqs.ReceiveMessageOutput) error {
	if message == nil {
		return fmt.Errorf("Blank message")
	}
	for _, m := range message.Messages {
		fmt.Printf("msgID = %+v, msgAttri = %+v, msgBody = %+v", m.MessageId, m.MessageAttributes, m.Body)
	}
	return nil
}

func Test_sendMsg(t *testing.T) {
	type args struct {
		sess     *session.Session
		queueURL *string
		msg      SqsMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendMsg(tt.args.sess, tt.args.queueURL, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("sendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendMsg(t *testing.T) {
	type args struct {
		sess      *session.Session
		queueName string
		msg       SqsMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{queueName: "testQ", msg: &SqsMessagePicTest{}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendMsg(s, tt.args.queueName, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("SendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetQueueURL(t *testing.T) {
	type args struct {
		sess  *session.Session
		queue *string
	}
	tests := []struct {
		name    string
		args    args
		want    *sqs.GetQueueUrlOutput
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetQueueURL(tt.args.sess, tt.args.queue)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetQueueURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetQueueURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMessages(t *testing.T) {
	type args struct {
		sess      *session.Session
		queueName string
		output    SqsMessage
		timeout   *int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{sess: s, queueName: "testQ", output: &SqsMessagePicTest{}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetMessages(tt.args.sess, tt.args.queueName, tt.args.output, tt.args.timeout); (err != nil) != tt.wantErr {
				t.Errorf("GetMessages() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
