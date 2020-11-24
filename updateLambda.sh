#!/bin/bash
#
# Simple Script to do the build, pack and deploy to AWS using the Deploy IAM usr
# role configured via the cli 2.0 config
# 
# Note to make executable run chmod +x updateLambda.sh
#
GOOS=linux go build getregions.go
zip function.zip getregions

aws lambda update-function-code --function-name ASGS_01 --zip-file fileb://function.zip 
#aws lambda update-function-configuration --function-name DPA_01 --handler getregions --description 'Lambda function for calling dynamo, serverless mapping app'