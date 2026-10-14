package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/fx"

	"github.com/GabrielRodrigues-dev/backend_go/internal/usecase"
)

// SQSMessageBody modela exatamente o formato de mensagem exigido no desafio.
type SQSMessageBody struct {
	MessageId string `json:"messageId"`
	Data      struct {
		WalletId              string `json:"walletId"`
		ExternalTransactionId string `json:"externalTransactionId"`
		IdempotencyKey        string `json:"idempotencyKey"`
		Money                 struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"money"`
	} `json:"data"`
}

type SQSWorker struct {
	client     *sqs.Client
	queueURL   string
	processBet *usecase.ProcessBetUseCase
}

// NewSQSWorker recebe o cliente AWS e o UseCase injetados automaticamente pelo Uber Fx
func NewSQSWorker(client *sqs.Client, processBet *usecase.ProcessBetUseCase) *SQSWorker {
	return &SQSWorker{
		client:     client,
		queueURL:   "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/wager-transactions.fifo",
		processBet: processBet,
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
				var body SQSMessageBody
				if err := json.Unmarshal([]byte(*msg.Body), &body); err != nil {
					fmt.Printf("Erro ao ler JSON do SQS: %v\n", err)
					// Se o JSON for inválido, não apaga a mensagem.
					// Ela vai para a Dead Letter Queue (DLQ) após algumas tentativas.
					continue
				}

				fmt.Printf("\n>> [SQS WORKER] Processando aposta %s para carteira %s\n", body.Data.ExternalTransactionId, body.Data.WalletId)

				// 1. Mapeia a mensagem da fila para o formato do nosso Cérebro Central
				input := usecase.ProcessBetInput{
					WalletID:       body.Data.WalletId,
					TransactionID:  body.Data.ExternalTransactionId,
					Amount:         body.Data.Money.Amount,
					Currency:       body.Data.Money.Currency,
					IdempotencyKey: body.Data.IdempotencyKey, // A fila também usa Idempotência!
				}

				// 2. Executa as regras de negócio!
				err := w.processBet.Execute(ctx, input)
				if err != nil {
					fmt.Printf(">> [SQS WORKER] Operação rejeitada: %v\n", err)
				} else {
					fmt.Printf(">> [SQS WORKER] SUCESSO! Aposta salva no banco de dados.\n")
				}

				// 3. O desafio exige que a mensagem seja apagada apenas APÓS o tratamento no banco.
				w.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(w.queueURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}
