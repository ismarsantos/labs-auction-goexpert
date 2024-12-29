# Guia de Desenvolvimento

Este documento explica como subir a aplicação via Docker.

---

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) instalado
- [Docker Compose](https://docs.docker.com/compose/install/) instalado
- Arquivo `.env` configurado, caso existam variáveis de ambiente (ex.: `AUCTION_DURATION_MINUTES`, `AUCTION_CHECK_INTERVAL_SECONDS`, etc.).  
- Verifique no `docker-compose.yaml` qual o caminho esperado para esse arquivo (por exemplo, `cmd/auction/.env`).

---

## Passo a Passo

1. **Clone** o repositório:

   ```bash
   git clone https://github.com/devfullcycle/labs-auction-goexpert.git
   cd labs-auction-goexpert
   ```

2. **Crie (ou confirme que exista) o arquivo `.env` no local indicado pelo `docker-compose.yaml`**. Por exemplo, se ele estiver em `cmd/auction/.env`, abra/crie esse arquivo e configure as variáveis necessárias:

   ```bash
   AUCTION_DURATION_MINUTES=5
   AUCTION_CHECK_INTERVAL_SECONDS=10
   MONGO_URI=mongodb://mongodb:27017
   ```

3. **Construa e suba os contêineres com o Docker Compose:**

   ```bash
   docker-compose up --build
   ```

   Isso iniciará:
   - Um container MongoDB (definido em `docker-compose.yaml`).
   - Um container Go (definido em `Dockerfile`), que irá rodar a aplicação.

   Aguarde até que os contêineres estejam em execução. A aplicação ficará acessível em [http://localhost:8080](http://localhost:8080) (ou na porta definida no `docker-compose.yaml`).

4. **Para encerrar os contêineres, use:**

   ```bash
   docker-compose down
   ```

---

## Executando Testes

Caso deseje executar testes dentro do ambiente containerizado (e eles dependam do MongoDB ou de outros serviços), você pode:

1. **Subir apenas o MongoDB em modo desacoplado:**

   ```bash
   docker-compose up -d mongodb
   ```

2. **Rodar um container específico para testes, se existir um `Dockerfile` ou `docker-compose` voltado para isso.**

   Exemplo (hipotético):

   ```bash
   docker-compose run --rm app go test ./... -v
   ```

   (Ajuste conforme sua configuração.)

   *(Se você quiser executar testes localmente via `go test`, sem usar Docker, consulte outro guia ou faça as devidas adaptações.)*

---

## Observações

- Toda a lógica de variáveis de ambiente (ex.: `AUCTION_DURATION_MINUTES`, `AUCTION_CHECK_INTERVAL_SECONDS`, etc.) é lida pelo container Go em tempo de execução. Certifique-se de que o arquivo `.env` esteja no local correto, conforme indicado pelo `docker-compose.yaml`.
- Caso deseje alterar portas, nomes de contêineres ou caminhos de volume, ajuste o `docker-compose.yaml` conforme necessário.
- Para logs e diagnósticos, verifique a saída do terminal após executar `docker-compose up --build`.


