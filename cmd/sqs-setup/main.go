package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func main() {
	ctx := context.TODO()

	// Configura o cliente apontando para o LocalStack (porta 4566)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
	)
	if err != nil {
		log.Fatalf("Erro ao carregar config AWS: %v", err)
	}

	client := sqs.NewFromConfig(cfg)

	// Criação da DLQ FIFO
	dlqName := "wager-transactions-dlq.fifo"
	dlqRes, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(dlqName),
		Attributes: map[string]string{
			"FifoQueue": "true",
		},
	})
	if err != nil {
		log.Printf("Aviso ao criar DLQ: %v", err)
	} else {
		log.Printf("DLQ criada com sucesso: %s", *dlqRes.QueueUrl)
	}

	// Precisamos pegar o ARN da DLQ para configurar o redrive policy.
	// CORREÇÃO AQUI: Usando types.QueueAttributeName
	dlqAttr, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       dlqRes.QueueUrl,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	})
	var dlqArn string
	if err == nil {
		dlqArn = dlqAttr.Attributes["QueueArn"]
	}

	// Criação da Fila Principal FIFO com Redrive Policy para a DLQ
	queueName := "wager-transactions.fifo"
	redrivePolicy := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":"3"}`, dlqArn)

	queueRes, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"FifoQueue":                 "true",
			"ContentBasedDeduplication": "true",
			"RedrivePolicy":             redrivePolicy,
		},
	})
	if err != nil {
		log.Fatalf("Erro ao criar fila SQS principal: %v", err)
	}

	log.Printf("Fila SQS principal criada com sucesso: %s\n", *queueRes.QueueUrl)
}
