package main

import (
	"encoding/json"
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

// setup - This is why the dynamoDB api sucks with golang. I'm dying in a sea of Attribute Values. Gah!
func setup() dynamodb.BatchGetItemOutput {

	bgio := dynamodb.BatchGetItemOutput{
		Responses: map[string][]map[string]*dynamodb.AttributeValue{

			"testtable01": []map[string]*dynamodb.AttributeValue{
				{
					"RegionId": &dynamodb.AttributeValue{
						S: aws.String("115"),
					},
					"LevelType": &dynamodb.AttributeValue{
						S: aws.String("G02"),
					},
					"KVPairs": &dynamodb.AttributeValue{
						M: map[string]*dynamodb.AttributeValue{
							"Average_household_size":        &dynamodb.AttributeValue{N: aws.String("2.7")},
							"Average_num_psns_per_bedroom":  &dynamodb.AttributeValue{N: aws.String("0.7")},
							"Median_age_persons":            &dynamodb.AttributeValue{N: aws.String("36")},
							"Median_mortgage_repay_monthly": &dynamodb.AttributeValue{N: aws.String("2470")},
							"Median_rent_weekly":            &dynamodb.AttributeValue{N: aws.String("450")},
							"Median_tot_fam_inc_weekly":     &dynamodb.AttributeValue{N: aws.String("2646")},
							"Median_tot_hhd_inc_weekly":     &dynamodb.AttributeValue{N: aws.String("2330")},
							"Median_tot_prsnl_inc_weekly":   &dynamodb.AttributeValue{N: aws.String("1132")},
							"SA1_7DIGITCODE_2016":           &dynamodb.AttributeValue{N: aws.String("1101101")}},
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

	mr0 := RegionRequest{
		RegionID:    "115",
		LevelType: "SA4",
	}

	mr1 := RegionRequest{
		RegionID:    "123",
		LevelType: "SA4",
	}

	mr := []RegionRequest{}

	mr = append(mr, mr0, mr1)

	//fmt.Print(mr)

	req := events.APIGatewayProxyRequest{}

	b, _ := json.Marshal(mr)
	req.Body = string(b)

	x, _ := d.HandleRequest(req)

	fmt.Print("\n----\n")
	fmt.Println("x")
	fmt.Println(x)

}
