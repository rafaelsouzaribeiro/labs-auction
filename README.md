# Labs Auction - Fechamento Automático de Leilões (Mba Exercícios)

## Antes de começar

Crie (ou valide) o arquivo de ambiente em [cmd/auction/.env](cmd/auction/.env) com os parâmetros abaixo:

```dotenv
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4
AUCTION_INTERVAL=20s
AUCTION_DURATION=5s
MONGODB_URL=mongodb://root:root@mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
```

## Docker Compose (criar e subir ambiente)

Execute na raiz do projeto:

```bash
docker compose up -d --build
```

Para acompanhar logs:

```bash
docker compose logs -f
```

Para parar o ambiente:

```bash
docker compose down
```

## API Routes

### Auctions

#### Create Auction
**POST** `/auction`

```json
{
  "product_name": "Product Name",
  "category": "Category",
  "description": "Product description here",
  "condition": 0
}
```

> `condition`: `0` = New, `1` = Used, `2` = Refurbished

---

#### Find Auctions
**GET** `/auction?status=0&category=Electronics&productName=Phone`

| Query Param   | Type   | Description              |
|---------------|--------|--------------------------|
| `status`      | int    | `0` = Active, `1` = Closed |
| `category`    | string | Filter by category       |
| `productName` | string | Filter by product name   |

---

#### Find Auction by ID
**GET** `/auction/:auctionId`

---

#### Find Winning Bid by Auction ID
**GET** `/auction/:auctionId/winner`

---

### Bids

#### Create Bid
**POST** `/bid`

```json
{
  "user_id": "uuid-do-usuario",
  "auction_id": "uuid-do-leilao",
  "amount": 150.00
}
```

---

#### Find Bids by Auction ID
**GET** `/bid/:auctionId`

## Objetivo

Adicionar uma funcionalidade crítica ao sistema de leilões existente: o fechamento automático.

Atualmente, o projeto permite criar leilões e dar lances, mas o leilão nunca expira. A missão é utilizar Goroutines para garantir que o leilão seja encerrado automaticamente após um tempo pré-definido.

## Requisitos técnicos

### 1) Ajuste no fluxo de criação do leilão

- Modificar o processo de criação de leilão para incluir o agendamento do fechamento.

### 2) Configuração de tempo

- Criar uma função (ou ajustar as existentes) para determinar a duração do leilão com base em variáveis de ambiente (ex.: `AUCTION_DURATION`).

### 3) Processamento em background (Goroutine)

- Iniciar uma Goroutine assim que um leilão for criado.
- A rotina deve monitorar o tempo de duração.
- Quando o prazo expirar, deve atualizar o banco de dados alterando o status do leilão para `Closed`.

## Critérios de aceite

- Um leilão criado com duração configurada deve ser fechado automaticamente após o tempo definido.
- O status final no banco deve ser `Closed`.
- O comportamento deve ser consistente para múltiplos leilões criados em sequência.

## Observações de implementação

- Preferir uso de `time.Duration` com `time.ParseDuration` para tratar o valor de `AUCTION_DURATION`.
- Garantir que a atualização de status seja idempotente (evitar fechar novamente um leilão já encerrado).
- Tratar erros da rotina de fechamento com logging apropriado.