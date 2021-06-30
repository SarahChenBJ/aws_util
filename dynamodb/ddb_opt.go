package dynamodb

import (
	"errors"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

/*GetKeySchema finds out all Keys with key_type:"HASH;" or key_type:"RANGE;" from a struct */
func getKeySchema(i interface{}) []*dynamodb.KeySchemaElement {
	t := reflect.TypeOf(i)
	var keys []*dynamodb.KeySchemaElement

	for i := 0; i < t.NumField(); i++ {

		field := t.Field(i)

		//name of field:
		//json tag comes first than the variable name
		attrName := getFieldName(field)

		if v, ok := field.Tag.Lookup(TagKey); ok {
			if strings.Contains(v, TagHash) {

				keys = append(keys, &dynamodb.KeySchemaElement{
					AttributeName: &attrName,
					KeyType:       aws.String(TagHash),
				})

			} else if strings.Contains(v, TagRange) {

				keys = append(keys, &dynamodb.KeySchemaElement{
					AttributeName: &attrName,
					KeyType:       aws.String(TagRange),
				})

			}
		}

	}
	return keys
}
func getAttriDef(i interface{}) []*dynamodb.AttributeDefinition {
	t := reflect.TypeOf(i)
	var dfs []*dynamodb.AttributeDefinition

	for i := 0; i < t.NumField(); i++ {

		field := t.Field(i)

		//name of field:
		//json tag comes first than the variable name
		n := getFieldName(field)

		//type of field:
		t := getDDBType(field)

		dfs = append(dfs, &dynamodb.AttributeDefinition{
			AttributeName: aws.String(n),
			AttributeType: aws.String(t),
		})
	}
	return dfs
}

func getFieldName(field reflect.StructField) string {
	n := field.Name
	jsn, jsok := field.Tag.Lookup("json")
	if jsok && jsn != "" {
		return jsn
	}
	return n
}

func getDDBType(field reflect.StructField) string {
	switch field.Type.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Chan, reflect.Ptr, reflect.UnsafePointer, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return "N"
	case reflect.String:
		return "S"
	default:
		return "B"
	}
}

func (c *dynamoClient) CreateTable(t IDynamotable) error {
	_, err := c.svc.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions:  t.GetAttriDef(),
		KeySchema:             t.GetKeySchema(),
		ProvisionedThroughput: t.GetProvisionedThroughput(),
		TableName:             t.GetTableName(),
	})
	// snippet-end:[dynamodb.go.create_new_table.call]
	return err
}

func (c *dynamoClient) ListTables(t *IDynamotable) ([]*string, error) {
	result, err := c.svc.ListTables(&dynamodb.ListTablesInput{
		Limit: aws.Int64(ListTablesLimit),
	})
	if err != nil {
		return nil, err
	}
	return result.TableNames, nil
}

// func LoadItems() {

// }

func (c *dynamoClient) GetItem(t IDynamotable) ([]IDynamotable, error) {
	result, err := c.svc.GetItem(&dynamodb.GetItemInput{
		TableName: t.GetTableName(),
		Key:       t.ConvertKeys(),
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("Not result")
	}
	var items []IDynamotable
	err = dynamodbattribute.UnmarshalMap(result.Item, items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (c *dynamoClient) UpsertItem(t IDynamotable) error {
	expr, upExpr := t.GenUpdateExpression()
	input := &dynamodb.UpdateItemInput{
		TableName:                 t.GetTableName(),
		Key:                       t.ConvertKeys(),
		ExpressionAttributeValues: expr,
		UpdateExpression:          upExpr,
		ReturnValues:              aws.String("UPDATE_NEW"),
	}
	_, err := c.svc.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (c *dynamoClient) AddItem(t IDynamotable) error {
	_, err := c.svc.PutItem(&dynamodb.PutItemInput{
		Item:      t.ConvertKeys(),
		TableName: t.GetTableName(),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *dynamoClient) DeleteItem(t IDynamotable) error {
	input := &dynamodb.DeleteItemInput{
		Key:       t.ConvertKeys(),
		TableName: t.GetTableName(),
	}
	_, err := c.svc.DeleteItem(input)

	if err != nil {
		return err
	}
	return nil
}
