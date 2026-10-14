package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/fx"
)

type SQSWorker struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSWorker(client *sqs.Client) *SQSWorker {
	return &SQSWorker{
		client: client,
		// O URL que o LocalStack gerou quando executou o script de setup
		queueURL: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/wager-transactions.fifo",
	}
}

// Registra o processo paralelo no Uber Fx
func StartWorker(lc fx.Lifecycle, worker *SQSWorker) {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			fmt.Println("Iniciando SQS Worker...")
			go worker.consumeLoop(ctx)
			return nil
		},
		OnStop: func(c context.Context) error {
			fmt.Println("Encerrando SQS Worker...")
			cancel() // Avisa o loop para parar a execução e libertar recursos
			return nil
		},
	})
}

func (w *SQSWorker) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return // Sai limpo quando o servidor é desligado
		default:
			// Pede até 10 mensagens à fila
			out, err := w.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(w.queueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     5, // Long polling de 5 segundos
			})

			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Erro ao ler SQS: %v", err)
				continue
			}

			for _, msg := range out.Messages {
				fmt.Printf(">> Mensagem recebida da Fila SQS: ID %s\n", *msg.MessageId)

				// O processamento de negócio será injetado aqui na próxima fase!

				w.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(w.queueURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}
