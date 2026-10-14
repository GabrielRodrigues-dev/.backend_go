# Decisões de Arquitetura

Este documento descreve as decisões de arquitetura e as abordagens técnicas adotadas no desafio.

## 1. Modelação do Dinheiro (Money)
O tipo `Money` foi implementado como um *value object* que armazena internamente os cêntimos em `int64`, evitando arredondamentos que ocorrem ao usar tipos de ponto flutuante (`float32`/`float64`). Foi adotada uma política de validação estrita, suportando apenas duas casas decimais, para garantir precisão e conformidade com as exigências.

## 2. Padrão de Transações e Concorrência
O sistema de persistência baseia-se num Repositório onde é garantida a atomicidade através de transações SQL.
Utilizámos o padrão *Pessimistic Locking* (`FOR NO KEY UPDATE`) associado ao *Optimistic Locking* (verificação de versão no `UPDATE`), garantindo proteção robusta contra concorrência e *lost updates*, forçando a serialized isolation em disputas por uma mesma carteira.

## 3. Padrão Inbox e Outbox
- **Inbox:** Para assegurar que as mensagens SQS (e as requisições HTTP) são processadas de forma idempotente, a tabela `inbox` foi integrada na mesma transação atómica que regista o ledger e altera o saldo.
- **Outbox:** Eventos são serializados de forma imutável e guardados na tabela `outbox` também durante a transação principal. Um *worker* secundário varre regularmente a base de dados para garantir que a sua publicação ocorrerá pelo menos uma vez, mesmo em cenários de falha.

## 4. Uber Fx
A aplicação adota injeção de dependências controlada pelo framework *Uber Fx*.
- Cada componente é injetado via `fx.Provide`.
- O ciclo de vida do servidor web e dos *workers* (SQS e Outbox) são orquestrados com o `fx.Lifecycle`. As operações suportam *graceful shutdown*, finalizando processos que estejam a correr antes do encerramento completo do processo pai.

## 5. Limitações e Trabalhos Não Concluídos
Foram implementados o essencial da API HTTP (`POST /bet` e `POST /wallets`) e as transações de `DEBIT`/`CREDIT`.
Rotas secundárias, testes unitários mais profundos e funcionalidades como reenvios em backoff (para referências pendentes) e reconciliação detalhada poderão ser explorados numa iteração subsequente.