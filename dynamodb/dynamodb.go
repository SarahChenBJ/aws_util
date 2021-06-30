package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	ListTablesLimit = 100
	TagKey          = "ddb"
	TagHash         = "HASH"
	TagRange        = "RANGE"
)

type dynamoClient struct {
	svc dynamodbiface.DynamoDBAPI
}

func NewDynamoClient(sess *session.Session) *dynamoClient {
	return &dynamoClient{
		svc: dynamodb.New(sess),
	}
}

//IDynamotable is the interface for all tables
type IDynamotable interface {
	GetKeySchema() []*dynamodb.KeySchemaElement
	GetAttriDef() []*dynamodb.AttributeDefinition

	GetProvisionedThroughput() *dynamodb.ProvisionedThroughput
	GetTableName() *string

	ConvertKeys() map[string]*dynamodb.AttributeValue
	ConvertValues() map[string]*dynamodb.AttributeValue
	ConvertAll() map[string]*dynamodb.AttributeValue

	GenUpdateExpression() (map[string]*dynamodb.AttributeValue, *string)
	AbsorbValues(av map[string]*dynamodb.AttributeValue) error
}
