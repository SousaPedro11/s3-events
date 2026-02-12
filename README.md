# s3-events — POC

Proof-of-concept project for processing S3 events using an AWS Lambda written in Go. The Lambda receives notifications about object creation and removal from S3 (via SNS) and emits structured JSON logs (event-sourcing) that can be consumed by another service to update metadata status in a database.

## Overview

- Infrastructure: Terraform (resources are under `infra/tf`) — S3, SNS, Lambda, IAM.
- Application: Go Lambda in `cmd/lambda` — the handler processes SNS messages that contain S3 Event records.
- Typical flow: backend generates presigned URL → front-end uploads object → backend stores metadata (pending) → S3 sends notification via SNS → Lambda emits a JSON command (e.g. set_status uploaded) → a consumer applies the change to the database.

## Architecture

- S3 -> SNS -> Lambda (Go handler)
- Lambda writes single-line JSON events to logs for downstream consumers (event-sourcing pattern)

## Prerequisites

- Go (version compatible with `go.mod`, e.g. 1.25+)
- Terraform (>= 1.4)
- AWS CLI configured with credentials and permissions
- zip utility

## Repository layout

- `infra/tf/` — Terraform configurations (separated by resource)
- `cmd/lambda/` — Go source for the Lambda and tests

## Important variables

- Define variables in `infra/tf/terraform.tfvars` such as `bucket_name`, `region`, `lambda_name`.
- Note: Terraform backend configuration is in `infra/tf/main.tf`. The backend is initialized before other variables are evaluated — do not rely on tfvars inside the backend block.

## Running locally with Terraform

1. Initialize Terraform:

```bash
cd infra/tf
terraform init
```

2. Create an execution plan:

```bash
terraform plan -out s3-events.tfplan
```

3. Apply the plan:

```bash
terraform apply "s3-events.tfplan"
```

Note: This POC includes a `local-exec` provisioner that builds the Lambda and produces a ZIP artifact under `infra/tf/build`. For production or CI pipelines, prefer building artifacts outside Terraform and using `source_code_hash` or use `package_type = "Image"` with ECR.

## Testing the Lambda locally

<!-- 1. Manual build and package (instead of using the Terraform provisioner):

```bash
# from the repository root
cd cmd/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/bootstrap .
chmod +x build/bootstrap
cd build
zip -r ../../infra/tf/build/s3-events-handler.zip bootstrap
``` -->

Run tests and generate coverage report:

```bash
cd cmd/lambda
go test -coverprofile cover.out ./...
go tool cover -html cover.out -o cover.html
```

## S3 events observed

By default the project listens to the S3 events that affect metadata status:

- `s3:ObjectCreated:*` — map to `uploaded`
- `s3:ObjectRemoved:*` — map to `deleted`

Other S3 events (replication, lifecycle, ACL changes) can be added but increase noise and costs. Use `prefix`/`suffix` filters to narrow the notifications.

## Bucket deletion notes (BucketNotEmpty)

If `terraform destroy` fails with `BucketNotEmpty` the bucket contains objects or object versions. To resolve:

- Empty the bucket manually via AWS CLI or Console before deletion
- For development-only buckets, set `force_destroy = true` in the `aws_s3_bucket` resource to let Terraform remove objects/versions automatically (destructive)

## Recommended improvements

- Move the build step into CI (e.g., GitHub Actions) and use `filebase64sha256` or container images in ECR
- Introduce SQS between S3 and Lambda for buffering, retries and a DLQ
- Implement a consumer that applies the JSON commands to the database in an idempotent way

<!-- ## License & Contact

POC project — internal use. For questions or help creating CI pipelines, contact the repository maintainer. -->

---

This README was generated as a starting point for the POC.

<!-- # s3-events — POC

Projeto POC para processar eventos S3 com uma Lambda escrita em Go. Recebe notificações de criação/remoção de objetos no S3 (via SNS) e emite logs em formato JSON (event-sourcing) que podem acionar outra função para atualizar status de metadados em banco.

## Visão geral

- Infra: Terraform — cria/usa S3, SNS, Lambda e IAM (arquivos em `infra/tf`).
- App: Lambda em Go em `cmd/lambda` — handler processa mensagens SNS que contêm S3 Events.
- Fluxo Padrão: Backend gera presigned URL → Front faz upload → backend salva metadados (pending) → S3 notifica via SNS → Lambda registra evento JSON (command set_status uploaded) → outro consumidor aplica mudança no banco.

## Arquitetura

- S3 -> SNS -> Lambda (handler em Go)
- Logs estruturados (JSON) na Lambda para event-sourcing

## Pré-requisitos

- Go (versão compatível com `go.mod` — p.ex. 1.25+)
- Terraform (>= 1.4)
- AWS CLI configurado (credenciais com permissões necessárias)
- zip

## Estrutura do repositório

- `infra/tf/` — código Terraform por recurso
- `cmd/lambda/` — código-fonte Go da Lambda e testes

## Variáveis importantes

- `terraform.tfvars` (em `infra/tf`) contém valores como `bucket_name`, `region`, `lambda_name`.
- O backend do Terraform (state) é configurado em `infra/tf/main.tf` — não utilize variáveis dinamicamente dentro do bloco `backend`.

## Como rodar localmente (Terraform)

1. Inicialize o Terraform:

```bash
cd infra/tf
terraform init
```

2. Planejar (gera o plano e, se houver provisioner local, pode executar builds locais):

```bash
terraform plan -out s3-events.tfplan
```

3. Aplicar (aplica as mudanças e faz o deploy):

```bash
terraform apply "s3-events.tfplan"
```

Observação: este POC inclui um provisioner/local-exec que compila a Lambda e cria o ZIP no diretório `infra/tf/build`. Em ambientes CI é recomendado criar o artefato fora do Terraform e passar `source_code_hash` ou usar `package_type = "Image"` com ECR.

## Como construir e testar a Lambda (local)

1. Build e empacotar manualmente (se preferir não usar o provisioner):

```bash
# a partir do diretório raiz do repo
cd cmd/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/bootstrap .
chmod +x build/bootstrap
cd build
zip -r ../../infra/tf/build/s3-events-handler.zip bootstrap
```

2. Testes Go:

```bash
cd cmd/lambda
go test -coverprofile cover.out ./...
go tool cover -html cover.out -o cover.html
```

## Eventos S3 observados

Por padrão o projeto usa apenas os eventos que alteram status do metadata:

- `s3:ObjectCreated:*` — marcar `uploaded`
- `s3:ObjectRemoved:*` — marcar `deleted`

Outros eventos podem ser adicionados, porém aumentam o ruído e o custo (ex.: replication, lifecycle, acl). Recomendamos filtros `prefix`/`suffix` para limitar a relevância.

## Notas sobre exclusão de buckets (BucketNotEmpty)

Se você tentar `terraform destroy` e receber `BucketNotEmpty`, o bucket contém objetos (ou versões). Opções:

- Esvaziar manualmente (CLI ou Console) antes de deletar
- Se for bucket gerenciado e de desenvolvimento, `force_destroy = true` no recurso `aws_s3_bucket` fará o provider apagar objetos/versões automaticamente (cuidado: apaga tudo)

## Melhorias recomendadas

- Mover o build para CI (GitHub Actions/GitLab CI) e usar `filebase64sha256` ou imagem em ECR
- Inserir SQS entre S3 e Lambda para backpressure, retries e DLQ
- Adicionar um consumidor que aplique os comandos JSON no banco (idempotente)

## Licença & Contato

Projeto POC — uso interno. Para dúvidas, mudar código ou criar pipelines CI, entre em contato com o mantenedor do repo.

---

README gerado automaticamente como ponto de partida para o POC. -->
