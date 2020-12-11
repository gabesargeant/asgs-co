package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockGetItems struct {
	dynamodbiface.DynamoDBAPI
	GetItemResponse      dynamodb.GetItemOutput
	GetBatchItemResponse dynamodb.BatchGetItemOutput
}

func (d mockGetItems) GetItem(in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	fmt.Println("Mock!")
	return &d.GetItemResponse, nil
}

func (d mockGetItems) BatchGetItem(in *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	fmt.Println("Mock! batch Item response")

	bgio := setup()
	mgi := mockGetItems{
		GetBatchItemResponse: bgio,
	}

	return &mgi.GetBatchItemResponse, nil

}
//This is a valid response object.
//{"Nodes":[{"RegionID":"101","RegionName":"Capital Region","LevelType":"SA4","LevelIDName":"SA4_CODE_2016","ParentRegions":{"1":{"RegionID":"1","RegionName":"New South Wales","LevelType":"STATE"}},"ChildRegions":{"10102":{"RegionID":"10102","RegionName":"Queanbeyan","LevelType":"SA3"},"10103":{"RegionID":"10103","RegionName":"Snowy Mountains","LevelType":"SA3"},"10104":{"RegionID":"10104","RegionName":"South Coast","LevelType":"SA3"},"10105":{"RegionID":"10105","RegionName":"Goulburn - Mulwaree","LevelType":"SA3"},"10106":{"RegionID":"10106","RegionName":"Young - Yass","LevelType":"SA3"}}}]}

// setup - This is why the dynamoDB api sucks with golang. I'm dying in a sea of Attribute Values. Gah!
func setup() dynamodb.BatchGetItemOutput {

	bgio := dynamodb.BatchGetItemOutput{
		Responses: map[string][]map[string]*dynamodb.AttributeValue{

			"78978": []map[string]*dynamodb.AttributeValue{
				{
					"RegionId": &dynamodb.AttributeValue{
						S: aws.String("115"),
					},
					"LevelType": &dynamodb.AttributeValue{
						S: aws.String("SA4"),
					},
					"LevelIDName": &dynamodb.AttributeValue{
						S: aws.String("SA4_CODE_2016"),
					},
					"RegionName": &dynamodb.AttributeValue{
						S: aws.String("entral Coast"),
					},
					"ParentRegions": &dynamodb.AttributeValue{
							M: map[string]*dynamodb.AttributeValue{
							"123456" : {
						M: map[string]*dynamodb.AttributeValue{
							"RegionId": &dynamodb.AttributeValue{
								S: aws.String("1"),
							},
							"LevelType": &dynamodb.AttributeValue{
								S: aws.String("STATE"),
							},
							"RegionName": &dynamodb.AttributeValue{
								S: aws.String("New South Wales"),
							},
						},
					},},},
					"ChildRegions": &dynamodb.AttributeValue{
						M: map[string]*dynamodb.AttributeValue{
							"123456" : {
							M: map[string]*dynamodb.AttributeValue{
							"RegionId": &dynamodb.AttributeValue{
								S: aws.String("10202"),
							},
							"LevelType": &dynamodb.AttributeValue{
								S: aws.String("SA3"),
							},
							"RegionName": &dynamodb.AttributeValue{
								S: aws.String("Wyong"),
							},
						},},},
					},
				},
			},
		},
	}
	return bgio
}

// TestHandleRequest is the happy path test of the dynamodb BatchGetItem call for lambda getregions.go
func TestHandleRequest(t *testing.T) {

	m := mockGetItems{
		GetItemResponse: dynamodb.GetItemOutput{},
	}

	d := Dependencies{
		ddb:     m,
		tableID: "testTable",
	}

	mr0 := make(map[string]string)

	mr0["lvl"] = "SA4";
	//mr0["rgn"] = "101"

	//fmt.Print(mr)

	req := events.APIGatewayV2HTTPRequest{}

	
	req.QueryStringParameters = mr0

	x, _ := d.HandleRequest(req)

	fmt.Print("\n----\n")
	fmt.Println("x")
	fmt.Println(x)

}
