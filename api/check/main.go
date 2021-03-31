package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"context"
	"io/ioutil"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

type MailInfoData struct {
	To      string `json:"to"`
	Date    string `json:"date"`
	File    string `json:"file"`
	From    string `json:"from"`
	Subject string `json:"subject"`
}

type APIResponse struct {
	Message string         `json:"message"`
	Data    []MailInfoData `json:"data"`
}

type Response events.APIGatewayProxyResponse

var s3Client *s3.Client
var dynamodbClient *dynamodb.Client

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var jsonBytes []byte
	var err error
	d := make(map[string]string)
	json.Unmarshal([]byte(request.Body), &d)
	if v, ok := d["action"]; ok {
		switch v {
		case "getbody" :
			// check access token if you need
			if name, ok := d["name"]; ok {
				log.Println(name)
				res, err_ := getMailBody(ctx, name)
				if err_ != nil {
					err = err_
				} else {
					jsonBytes, _ = json.Marshal(APIResponse{Message: res, Data: []MailInfoData{}})
				}
			}
		case "getlist" :
			// check access token if you need
			res, err_ := getMailList(ctx)
			if err_ != nil {
				err = err_
			} else {
				jsonBytes, _ = json.Marshal(APIResponse{Message: "success", Data: res})
			}
		}
	}
	log.Print(request.RequestContext.Identity.SourceIP)
	if err != nil {
		log.Print(err)
		jsonBytes, _ = json.Marshal(APIResponse{Message: fmt.Sprint(err), Data: []MailInfoData{}})
		return Response{
			StatusCode: 500,
			Body: string(jsonBytes),
		}, nil
	}
	return Response {
		StatusCode: 200,
		Body: string(jsonBytes),
	}, nil
}

func scan(ctx context.Context, filt expression.ConditionBuilder, proj expression.ProjectionBuilder)(*dynamodb.ScanOutput, error)  {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		return nil, err
	}
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(os.Getenv("TABLE_NAME")),
	}
	res, err := dynamodbClient.Scan(ctx, input)
	return res, err
}

func getMailList(ctx context.Context)([]MailInfoData, error)  {
	var mailInfoList []MailInfoData
	// Filter by Destination
	filt := expression.NotEqual(expression.Name("to"), expression.Value(" "))
	// filt := expression.Equal(expression.Name("to"), expression.Value("UserName"))
	proj := expression.NamesList(expression.Name("to"), expression.Name("date"), expression.Name("file"), expression.Name("from"), expression.Name("subject"))
	res, err := scan(ctx, filt, proj)
	if err != nil {
		return mailInfoList, err
	}
	for _, i := range res.Items {
		item := MailInfoData{}
		err = attributevalue.UnmarshalMap(i, &item)
		if err != nil {
			return nil, err
		}
		mailInfoList = append(mailInfoList, item)
	}
	return mailInfoList, nil
}

func getMailBody(ctx context.Context, objectKey string)(string, error) {
	if len(objectKey) == 0 {
		return "", fmt.Errorf("Error: %s", "No ObjectKey.")
	}
	if s3Client == nil {
		s3Client = s3.NewFromConfig(getConfig(ctx))
	}
	input := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(objectKey),
	}
	res, err := s3Client.GetObject(ctx, input)
	if err != nil {
		return "", err
	}

	rc := res.Body
	defer rc.Close()
	tmpData, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if len(string(tmpData)) < 1 {
		return "", fmt.Errorf("Error: %s", "Empty Mail Body.")
	}

	// check user name if you need
	/*
	destination := ""
	for _, w := range regexp.MustCompile("[\n]").Split(string(tmpData), -1) {
		if strings.HasPrefix(w, "To") {
			destination = w[4:]
			break
		}
	}
	if destination != "UserName" {
		return "", fmt.Errorf("Error: %s", "Invalid Access.")
	}
	*/

	strNormalized := regexp.MustCompile("\r\n").ReplaceAllString(string(tmpData), "\n")
	rawMailLines := regexp.MustCompile(`\n\s*\n`).Split(strNormalized, -1)
	if len(rawMailLines) < 2 {
		return "", fmt.Errorf("Error: %s", "Empty Mail Body.")
	}
	return rawMailLines[1], nil
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		log.Print(err)
	}
	return cfg
}

func main() {
	lambda.Start(HandleRequest)
}
