package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
)

//Arguments for the program
type Arguments struct {
	DynamoDBTableName *string
	InputFile         *string
	OutputDir         *string
}

//Global Variable!
var outfolder string

func main() {
	//establish args.
	//fetch them//
	fmt.Println("Start")

	args := setArgs()
	flag.Parse()

	if *args.InputFile == "" {
		fmt.Print("No input specified.")
		os.Exit(9)
	}

	inputFile := getFile(*args.InputFile)

	outfolder = *args.OutputDir

	createOutDir(outfolder)

	mp := readCSV(inputFile)

	//summarizeRegions(mp)

	if *args.DynamoDBTableName != "" {
		pushToDatabase(*args.DynamoDBTableName, mp)
		fmt.Println("Sent : ", recordsSent, " to Dynamo")
	
	}else {
		fmt.Println("Skipping push to DB")
	}

}

func readCSV(file *os.File) map[string]AsgsRegionNode {

	br := bufio.NewReader(file)
	r := csv.NewReader(br)

	firstLine, err := r.Read()
	if err != nil {
		panic(err)
	}

	headMap := getHeaderMap(firstLine)

	buildNodes(headMap, r)

	mp := mergeLevels()

	return mp

}

func getFile(file string) *os.File {

	fmt.Printf("Attempting to read %s \n", file)
	f, err := os.Open(file)

	if err != nil {
		panic(err)
	}
	return f

}

func setArgs() Arguments {
	a := Arguments{}
	a.DynamoDBTableName = flag.String("n", "", "Name of the DynamoDB Table")
	a.InputFile = flag.String("i", "cat.csv", "Input File for building tree")
	a.OutputDir = flag.String("o", "./out/", "Output folder, if not set defaults to pwd ./out/ .")
	return a
}
