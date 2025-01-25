# Labs Go Expert - Concorrência com Golang - Leilão

## Objetivo

Adicionar uma nova funcionalidade ao projeto já existente para o leilão fechar automaticamente a partir de um tempo definido.

---

## Requisitos do Sistema

1. Uma função que irá calcular o tempo do leilão, baseado em parâmetros previamente definidos em variáveis de ambiente;

2. Uma nova go routine que validará a existência de um leilão (auction) vencido (que o tempo já se esgotou) e que deverá realizar o update, fechando o leilão (auction);

3. Um teste para validar se o fechamento está acontecendo de forma automatizada;

## Executando o Sistema

1. Certifique-se de ter o Docker, Docker Compose e Go instalados.
2. Clone o repositório e acesse o diretório raiz.
3. Execute o comando:

    ```bash
    docker compose up --build
    ```
4. O serviço estará disponível em: `http://localhost:8080`

## Testando o Sistema

### API

- **Endpoint**: `POST /auction`
- **Create Auction**:

    ```bash
    curl -X POST http://localhost:8080/auction  \
         -H 'Content-Type: application/json' \
         -d '{"product_name": "Product 1", "category": "Category 1", "description": "Description 1"}'
    ```

- **Endpoint**: `POST /bid`
- **Create Bid**:

    ```bash
    curl -X POST http://localhost:8080/bid  \
         -H 'Content-Type: application/json' \
         -d '{"user_id": "your_user_id", "your_auction_id": "Category 1", "amount": 100}'
    ```

### Teste Unitário
Execute o comando no diretório raiz:

```bash
cd internal/infra/database/auction && go test
```
