package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// A mensagem em JSON exigida pelo desafio.
	messageBody := `{
		"messageId": "sqs-msg-999",
		"type": "WagerTransactionRequested",
		"data": {
		  "providerId": "provider-a",
		  "externalTransactionId": "tx-sqs-001",
		  "idempotencyKey": "provider-a:tx-sqs-001",
		  "playerId": "player-123",
		  "walletId": "wallet-teste-1791925236121841100",
		  "roundId": "round-987",
		  "gameId": "fortune-chimp",
		  "kind": "BET",
		  "money": { "amount": "10.00", "currency": "BRL" }
		}
	  }`

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{URL: "http://localhost:4566", SigningRegion: "us-east-1"}, nil
	})

	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
	)

	client := sqs.NewFromConfig(cfg)
	queueURL := "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/wager-transactions.fifo"

	// Como é uma fila FIFO, precisamos do MessageGroupId
	groupID := "betting-group"

	_, err := client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:       aws.String(queueURL),
		MessageBody:    aws.String(messageBody),
		MessageGroupId: aws.String(groupID),
	})

	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}

	fmt.Println("Mensagem enviada com sucesso para a fila SQS!")
}
