package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

//AsgsRegionNode One region in all the regions, Let the nodes begin!
// This is a doubly linked node. ie it points up and down.
//Maps not arrays for the pointers, this to avoid duplicates
type AsgsRegionNode struct {
	RegionID        string                  `json:"RegionID,omitempty"`
	RegionName      string                  `json:"RegionName,omitempty"`
	LevelType       string                  `json:"LevelType,omitempty"`
	LevelIDName     string                  `json:"LevelIDName,omitempty"`
	ParentRegions   map[string]ParentRegion `json:"ParentRegions,omitempty"`
	ChildRegions    map[string]ChildRegion  `json:"ChildRegions,omitempty"`
}

//ChildRegion The output child of an Asgs Region Node
type ChildRegion struct {
	RegionID   string `json:"RegionID,omitempty"`
	RegionName string `json:"RegionName,omitempty"`
	LevelType  string `json:"LevelType,omitempty"`
}

//ParentRegion the output parent region of a ASGS region.
type ParentRegion struct {
	RegionID   string `json:"RegionID,omitempty"`
	RegionName string `json:"RegionName,omitempty"`
	LevelType  string `json:"LevelType,omitempty"`
}

// RegionNodeResponse array of nodes.
type RegionNodeResponse struct {
	Errors   []string          `json:"Errors,omitempty"`
	AsgsRegionNode  []AsgsRegionNode         `json:"Nodes"`	
}

//RegionRequest request. A region id and the level of that region. 
type RegionRequest struct {
	RegionID  string `json:"RegionID"`
	LevelType string `json:"PartitionID"`
}

var levelSequence = []string{
	"MB",
	"SA1",
	"SA2",
	"SA3",
	"SA4",
	"STATE",
	"AUS",
	"LGA",
	"POA",
	"SSC",
}

// Dependencies - Pointer Receiver based dependency injection
type Dependencies struct {
	ddb     dynamodbiface.DynamoDBAPI
	tableID string
}

// HandleRequest Main entry point for the Lambda
func (d *Dependencies) HandleRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var response events.APIGatewayProxyResponse
	var regionNodeResponse RegionNodeResponse

	var request []RegionRequest
	fmt.Println(request)
	//This will need to change to be the GET params.
	//TODO fix this.
	err := json.Unmarshal([]byte(req.Body), &request)

	if err != nil {
		fmt.Println("Error with unmarshalling request")
		fmt.Println(req.Body)

		response.StatusCode = 500
		s := []string{fmt.Sprint(err)}
		regionNodeResponse.Errors = s

		b, _ := json.Marshal(regionNodeResponse)
		response.Body = string(b)

		return response, errors.New("error with unmarshalling request")
	}

	//Validate Requests and trim long requests
	if len(request) >= 100 {
		fmt.Println("Trim request to 100 objects max")
		request = request[:99]

	}

	// Request items from DB.
	db := d.ddb
	table := d.tableID

	regionNodeResponse = getBatchData(request, db, table)
	
	b, err := json.Marshal(regionNodeResponse)

	if err != nil {
		fmt.Println("error with marshalling request")
		response.StatusCode = 500
		s := []string{fmt.Sprint(err)}
		regionNodeResponse.Errors = s

	} else {
		response.Body = string(b)
		response.StatusCode = 200
	}

	//fmt.Print(response)
	//fmt.Print(response.Body)

	return response, nil
}

// main Establish Go session and call lambda start with pointer receiver.
func main() {

	d := Dependencies{
		ddb:     dynamodb.New(session.New()),
		tableID: os.Getenv("DYNAMOTABLE"),
	}

	lambda.Start(d.HandleRequest)
}

// getBrachData - Makes requests on dynamodb with a batch interface.
func getBatchData(requests []RegionRequest, ddb dynamodbiface.DynamoDBAPI, dbTable string) RegionNodeResponse {

	mapOfKeys := []map[string]*dynamodb.AttributeValue{}

	for _, request := range requests {

		mapOfKeys = append(mapOfKeys, map[string]*dynamodb.AttributeValue{
			"RegionID": {
				S: aws.String(request.RegionID),
			},
			"LevelType": {
				S: aws.String(request.LevelType),
			},
		})

	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			dbTable: {
				Keys: mapOfKeys,
			},
		},
	}

	batch, err := ddb.BatchGetItem(input)

	var errors []string

	if err != nil {
		errors = append(errors, processErrors(err))
	}

	var fullResults []AsgsRegionNode

	for _, response := range batch.Responses {
		for _, item := range response {

			var result AsgsRegionNode
			err := dynamodbattribute.UnmarshalMap(item, &result)

			if err != nil {

				errorMsg := fmt.Sprint(err)
				errors = append(errors, errorMsg)

			}
			fullResults = append(fullResults, result)
		}
	}

	regionNodeResponse := RegionNodeResponse{		
		AsgsRegionNode: fullResults,
		Errors:  errors,
	}

	return regionNodeResponse
}

func processErrors(err error) string {
	var errorMessage string
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				errorMessage = fmt.Sprint(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				errorMessage = fmt.Sprint(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				errorMessage = fmt.Sprint(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				errorMessage = fmt.Sprint(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				errorMessage = fmt.Sprint(aerr.Error())
			}
		} else {
			errorMessage = fmt.Sprint(err.Error())
		}
	}
	return errorMessage
}