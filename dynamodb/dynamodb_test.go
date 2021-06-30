package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type Movie struct {
	Year   int     `json:"movie_year" ddb:"HASH"`
	Title  string  `json:"movie_title" ddb:"RANGE"`
	Rating float32 `json:"movie_rating"`
}

type TestStruct struct {
	IntVal     int     `json:"int_val" ddb:"int_val"`
	Int32Val   int32   `json:"int_32_val" ddb:"int_32_val"`
	Int64Val   int64   `json:"int_64_val" ddb:"int_64_val"`
	Float32Val float32 `json:"float_32_val" ddb:"float_32_val"`
	Float64Val float64 `json:"float_64_val" ddb:"float_64_val"`
	StringVal  string  `json:"string_val" ddb:"string_val"`
	StructVal  Movie   `json:"struct_val" ddb:"struct_val"`
	privateVal int
}

func (t *Movie) GetKeySchema() []*dynamodb.KeySchemaElement {
	// return []*dynamodb.KeySchemaElement{
	// 	{
	// 		AttributeName: aws.String("Year"),
	// 		KeyType:       aws.String("HASH"),
	// 	},
	// 	{
	// 		AttributeName: aws.String("Title"),
	// 		KeyType:       aws.String("RANGE"),
	// 	},
	// }
	return getKeySchema(t)
}

func (t *Movie) GetTableName() *string {
	return aws.String("movie")
}

func (t *Movie) GetProvisionedThroughput() *dynamodb.ProvisionedThroughput {
	return &dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  aws.Int64(10),
		WriteCapacityUnits: aws.Int64(10),
	}
}

func (t *Movie) GetAttriDef() []*dynamodb.AttributeDefinition {
	// return []*dynamodb.AttributeDefinition{
	// 	{
	// 		AttributeName: aws.String("Year"),
	// 		AttributeType: aws.String("N"),
	// 	},
	// 	{
	// 		AttributeName: aws.String("Title"),
	// 		AttributeType: aws.String("S"),
	// 	},
	// }
	return getAttriDef(t)
}

func (t *Movie) SetAttr(av map[string]*dynamodb.AttributeValue) error {
	err := dynamodbattribute.UnmarshalMap(av, t)
	return err
}

func Test_dynamoClient_CreateTable(t *testing.T) {
	type fields struct {
		svc dynamodbiface.DynamoDBAPI
	}
	type args struct {
		t IDynamotable
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "possitive", args: args{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &dynamoClient{
				svc: tt.fields.svc,
			}
			if err := c.CreateTable(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("dynamoClient.CreateTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ParseDDBStruct(t *testing.T) {
	sn, file := "TestStruct", "dynamodb_test.go"
	e := ParseDDBStruct(sn, file)
	if e != nil {
		t.Error(e)
	}
}

func Test_GenStubs(t *testing.T) {
	sn, file := "TestStruct", "dynamodb_test.go"
	e := GenStubs("github.com/aws-opt-go/utils/dynamodb.IDynamotable", ".", sn, file)
	t.Error(e)

	if e != nil {
	}
}

func Test_orchestra(t *testing.T) {
	sn, file := "TestStruct", "dynamodb_test.go"
	e1 := ParseDDBStruct(sn, file)
	e2 := GenStubs("github.com/aws-opt-go/utils/dynamodb.IDynamotable", ".", sn, file)

	if e1 != nil || e2 != nil {
		t.Error(e1, e2)
	}
}
