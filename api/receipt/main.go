package main

import (
	"os"
	"log"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

type MailInfoData struct {
	To      string `dynamodbav:"to"`
	Date    string `dynamodbav:"date"`
	File    string `dynamodbav:"file"`
	From    string `dynamodbav:"from"`
	Subject string `dynamodbav:"subject"`
}

type APIResponse struct {
	Message string `json:"message"`
}

var dynamodbClient *dynamodb.Client

const layout  string = "2006-01-02 15:04"
const layout2 string = "20060102150405"

func HandleRequest(ctx context.Context, sesEvent events.SimpleEmailEvent) error {
	for _, record := range sesEvent.Records {
		ses := record.SES
		err := putMailInfo(ctx, ses.Mail.CommonHeaders.To[0], ses.Mail.CommonHeaders.Date, ses.Mail.MessageID, ses.Mail.CommonHeaders.From[0], ses.Mail.CommonHeaders.Subject)
		if err != nil {
			log.Println(err)
			log.Printf("CommonHeaders %+v", ses.Mail.CommonHeaders)
		}
	}

	return nil
}

func put(ctx context.Context, tableName string, av map[string]dynamodbtypes.AttributeValue) error {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}
	_, err := dynamodbClient.PutItem(ctx, input)
	return err
}

func putMailInfo(ctx context.Context, to string, date string, file string, from string, subject string) error {
	item := MailInfoData {
		To: to,
		Date: date,
		File: file,
		From: from,
		Subject: subject,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	err = put(ctx, os.Getenv("TABLE_NAME"), av)
	if err != nil {
		return err
	}
	return nil
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
