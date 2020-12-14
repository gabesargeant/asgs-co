package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

//LevelName, Level Code
var levels = map[string]map[string]AsgsRegionNode{
	"MB":    {},
	"SA1":   {},
	"SA2":   {},
	"SA3":   {},
	"SA4":   {},
	"STATE": {},
	"AUS":   {},
	"LGA":   {},
	"POA":   {},
	"SSC":   {},
}

var totalRegions float64 = 0

func mergeLevels() map[string]AsgsRegionNode {

	mp := make(map[string]AsgsRegionNode)
	for k, v := range levels {

		fmt.Printf("Region Type: %s, size: %d \n", k, len(v))

		for kk, vv := range v {
			mp[kk] = vv
		}

	}
	totalRegions = float64(len(mp))
	fmt.Println("totalRegions", totalRegions )
	return mp
}

//The order of regions in this array is significant.
//I need to build the nodes from the MeshBlocks up.
//as each 'higher' region will require the lower to be there
//to link it. And each child gets linked it's parent when
//its parent gets its child.
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

var levelCodeMap = map[string]string{

	"MB":    "MB_CODE_2016",
	"SA1":   "SA1_MAINCODE_2016",
	"SA2":   "SA2_MAINCODE_2016",
	"SA3":   "SA3_CODE_2016",
	"SA4":   "SA4_CODE_2016",
	"STATE": "STATE_CODE_2016",
	"AUS":   "AUS_CODE_2016",
	"LGA":   "LGA_CODE_2020",
	"POA":   "POA_CODE_2016",
	"SSC":   "SSC_CODE_2016",
}

var levelNameMap = map[string]string{

	"MB":    "MB_CATEGORY_NAME_2016",
	"SA1":   "SA1_NAME_2016",
	"SA2":   "SA2_NAME_2016",
	"SA3":   "SA3_NAME_2016",
	"SA4":   "SA4_NAME_2016",
	"STATE": "STATE_NAME_2016",
	"AUS":   "AUS_NAME_2016",
	"LGA":   "LGA_NAME_2020",
	"POA":   "POA_NAME_2016",
	"SSC":   "SSC_NAME_2016",
}

//Australia,
//States
//SA4
//SA3
//SA2
//SA1
//MB -- LGA -- POA -- SSC
//ChildLevles Key = the current region, value = it's child region
var asgsChildLevel = map[string]string{
	"MB":    "",
	"SA1":   "MB",
	"SA2":   "SA1",
	"SA3":   "SA2",
	"SA4":   "SA3",
	"STATE": "SA4",
	"AUS":   "STATE",
	"LGA":   "MB",
	"POA":   "MB",
	"SSC":   "MB",
}

// //The many to many is a problem....hmm.
// var nonAsgsChildLevel = map[string][]string{
// 	"MB":    {},
// 	"STATE": {"LGA", "POA", "SSC"},
// 	"AUS":   {"STATE"},
// 	"LGA":   {"MB"},
// 	"POA":   {"MB"},
// 	"SSC":   {"MB"},
// }

var skipLevel = map[string]string{
	"MB":  "SKIP",
	"SA1": "SKIP",
}

var asgsRegionArray = []string{
	"MB_CODE_2016",
	"MB_CATEGORY_NAME_2016",
	"SA1_MAINCODE_2016",
	"SA1_NAME_2016",
	"SA2_MAINCODE_2016",
	"SA2_NAME_2016",
	"SA3_CODE_2016",
	"SA3_NAME_2016",
	"SA4_CODE_2016",
	"SA4_NAME_2016",
	"STATE_CODE_2016",
	"STATE_NAME_2016",
	"AUS_CODE_2016",
	"AUS_NAME_2016",
	"LGA_CODE_2020",
	"LGA_NAME_2020",
	"POA_CODE_2016",
	"POA_NAME_2016",
	"SSC_CODE_2016",
	"SSC_NAME_2016",
}

//getHeaderMap -
func getHeaderMap(firstLine []string) map[string]int {

	m := make(map[string]int)

	for i := 0; i < len(asgsRegionArray); i++ {

		for j := 0; j < len(firstLine); j++ {

			areaValue := asgsRegionArray[i]

			header := firstLine[j]

			if areaValue == header {
				m[areaValue] = j
			}

		}
	}

	return m
}

func buildNodes(headerMap map[string]int, r *csv.Reader) {

	//outter loop, read a row.
	for {
		row, err := r.Read()
		if err == io.EOF {
			//for the last record, write the buffer without the commaNewLine
			break
		}

		for _, currentLevel := range levelSequence {

			levelCode := levelCodeMap[currentLevel]
			levelName := levelNameMap[currentLevel]

			//instanceLevelCode
			iLevelCode := row[headerMap[levelCode]]
			iLevelName := row[headerMap[levelName]]

			region := levels[currentLevel][iLevelCode]

			if region.RegionID == "" {

				region.LevelIDName = levelCode
				region.LevelType = currentLevel
				region.RegionName = iLevelName
				region.RegionID = iLevelCode
				region.ChildRegions = make(map[string]ChildRegion)
				region.ParentRegions = make(map[string]ParentRegion)

			}

			//Add child element
			child := asgsChildLevel[currentLevel]

			if child == "" {
				levels[currentLevel][iLevelCode] = region
				continue
			}

			childLevelCode := levelCodeMap[child]

			childRegionCode := row[headerMap[childLevelCode]]

			childRegion := levels[child][childRegionCode]

			//Establish Relationships
			pr := ParentRegion{}
			pr.LevelType = region.LevelType
			pr.RegionName = region.RegionName
			pr.RegionID = region.RegionID

			childRegion.ParentRegions[region.RegionID] = pr
			cr := ChildRegion{}
			cr.LevelType = childRegion.LevelType
			cr.RegionName = childRegion.RegionName
			cr.RegionID = childRegion.RegionID
			region.ChildRegions[childRegion.RegionID] = cr

			//Set objects
			levels[currentLevel][iLevelCode] = region
			levels[child][childRegionCode] = childRegion

		}
	}
}

func summarizeRegions(regions map[string]AsgsRegionNode) {
	var i int = 0
	fmt.Println("starting region output build")
	for _, v := range regions {
		i++
		printRegion(v.RegionID, v)
		if i > 100 {
			break
		}

	}

}

func pushToDatabase(tableName string, mapNodes map[string]AsgsRegionNode) {

	
	// //one session for all uploads.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		//SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{Region: aws.String("ap-southeast-2")},
	}))

	remainder := len(mapNodes) % 25;
	finNodeArr := make([]AsgsRegionNode,remainder)

	nodeArr := make([]AsgsRegionNode,25)
	//fmt.Println(len(mapNodes))
	i := 0
	for _, v := range mapNodes {
		i++
		// fmt.Println(v);
		// fmt.Println(i);
		nodeArr[i%25] = v
		if (i % 25 == 0 && i > 0) || i == len(mapNodes) -1 {
			

//			fmt.Println(nodeArr)
			if i != len(mapNodes)-1 {
				br := getBatch(tableName, nodeArr)
				pushToDynamo(sess, br)
			}else{
				// final nodes
				//fmt.Println(nodeArr)
				//fmt.Println("Final Nodes ----")

				for j := 0; j < remainder; j++{
					finNodeArr[j]= nodeArr[j]
				
				}

				br := getBatch(tableName, nodeArr)
				pushToDynamo(sess, br)

				
			}
			
		
			
			
		}
		

	}

	for _, v := range failOverRequest {
		processFailOver(sess, v);
	}


}

var recordsSent float64 = 0;

func pushToDynamo(sess *session.Session, batchReq dynamodb.BatchWriteItemInput) {
	//return
	svc := dynamodb.New(sess)

	result, err := svc.BatchWriteItem(&batchReq)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

	}

	if len(result.UnprocessedItems) != 0 {
		fmt.Println(*result)

		fmt.Println("Some records no processed, pushing to failover.");
		
		bwi := dynamodb.BatchWriteItemInput{}
		bwi.SetRequestItems(result.UnprocessedItems)
		failOverRequest = append(failOverRequest, bwi)

		fmt.Println("Fail Over size:", len(failOverRequest))

	}

	recordsSent += 25.0

	
	x := math.Round((recordsSent/totalRegions) * 100)

	fmt.Println("Percentage done : ~ ", x, "%")
}

var failOverRequest = []dynamodb.BatchWriteItemInput{}

func processFailOver(sess *session.Session, batchReq dynamodb.BatchWriteItemInput){

	svc := dynamodb.New(sess)

	result, err := svc.BatchWriteItem(&batchReq)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

	}

	if len(result.UnprocessedItems) != 0 {
		fmt.Println("failover with fail over")
		fmt.Println("Printing out final nodes");

		for _, v := range result.UnprocessedItems {

			for i := 0; i < len(v); i++ {

				pr := v[i].PutRequest
				var node AsgsRegionNode
					err := dynamodbattribute.UnmarshalMap(pr.Item, &node)
					if err != nil {
						fmt.Println("can't unmarshal ", pr.Item);

					}
					printRegion(node.RegionID, node)
			}	
		}		
	}
}

//Push to dynamoDB table.
func getBatch(tableName string, nodeArr []AsgsRegionNode) dynamodb.BatchWriteItemInput {

	wrArr := []*dynamodb.WriteRequest{}
	//fmt.Println("node len ", len(nodeArr))
	for _, n := range nodeArr {
		//fmt.Println("node")
		//fmt.Println(nodeArr)

		av, err := dynamodbattribute.MarshalMap(n)
		//fmt.Println("attribute maps")
		//fmt.Println(av)
		//os.Exit(1)
		if err != nil {
			fmt.Println("Error with unmarhalling list of nodesets")
			fmt.Println(err.Error())
			os.Exit(1)

		}
		pr := dynamodb.PutRequest{}
		pr.SetItem(av)
		wr := dynamodb.WriteRequest{}
		wr.SetPutRequest(&pr)
		wrArr = append(wrArr, &wr)

	}

	wrMap := make(map[string][]*dynamodb.WriteRequest, 1)
	wrMap[tableName] = wrArr

	bwi := dynamodb.BatchWriteItemInput{}
	bwi.SetRequestItems(wrMap)

	return bwi

}

//Not needed
func createOutDir(outDir string) {

	err := os.Mkdir(outDir, 0777)
	if os.IsNotExist(err) {
		fmt.Printf("Error with creating output folder %s", err)
	}
}

//only for testing
func printRegion(id string, out AsgsRegionNode) {

	dataFile, err := os.Create(outfolder + id + ".json")
	bw := bufio.NewWriter(dataFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(9)
	}

	var jsonData []byte
	jsonData, err = json.MarshalIndent(out, "", "\t")

	if err != nil {
		fmt.Println(err)
		os.Exit(9)
	}
	bw.Write(jsonData)
	bw.Flush()
	dataFile.Close()

}
