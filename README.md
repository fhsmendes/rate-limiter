# rate-limiter
Desafio fullcycle - Rate Limiter

Objetivo: Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

Descrição: O objetivo deste desafio é criar um rate limiter em Go que possa ser utilizado para controlar o tráfego de requisições para um serviço web. O rate limiter deve ser capaz de limitar o número de requisições com base em dois critérios:

Endereço IP: O rate limiter deve restringir o número de requisições recebidas de um único endereço IP dentro de um intervalo de tempo definido.
Token de Acesso: O rate limiter deve também poderá limitar as requisições baseadas em um token de acesso único, permitindo diferentes limites de tempo de expiração para diferentes tokens. O Token deve ser informado no header no seguinte formato:
API_KEY: <TOKEN>
As configurações de limite do token de acesso devem se sobrepor as do IP. Ex: Se o limite por IP é de 10 req/s e a de um determinado token é de 100 req/s, o rate limiter deve utilizar as informações do token.
Requisitos:

O rate limiter deve poder trabalhar como um middleware que é injetado ao servidor web
O rate limiter deve permitir a configuração do número máximo de requisições permitidas por segundo.
O rate limiter deve ter ter a opção de escolher o tempo de bloqueio do IP ou do Token caso a quantidade de requisições tenha sido excedida.
As configurações de limite devem ser realizadas via variáveis de ambiente ou em um arquivo “.env” na pasta raiz.
Deve ser possível configurar o rate limiter tanto para limitação por IP quanto por token de acesso.
O sistema deve responder adequadamente quando o limite é excedido:
Código HTTP: 429
Mensagem: you have reached the maximum number of requests or actions allowed within a certain time frame
Todas as informações de "limiter” devem ser armazenadas e consultadas de um banco de dados Redis. Você pode utilizar docker-compose para subir o Redis.
Crie uma “strategy” que permita trocar facilmente o Redis por outro mecanismo de persistência.
A lógica do limiter deve estar separada do middleware.
Exemplos:

Limitação por IP: Suponha que o rate limiter esteja configurado para permitir no máximo 5 requisições por segundo por IP. Se o IP 192.168.1.1 enviar 6 requisições em um segundo, a sexta requisição deve ser bloqueada.
Limitação por Token: Se um token abc123 tiver um limite configurado de 10 requisições por segundo e enviar 11 requisições nesse intervalo, a décima primeira deve ser bloqueada.
Nos dois casos acima, as próximas requisições poderão ser realizadas somente quando o tempo total de expiração ocorrer. Ex: Se o tempo de expiração é de 5 minutos, determinado IP poderá realizar novas requisições somente após os 5 minutos.
Dicas:

Teste seu rate limiter sob diferentes condições de carga para garantir que ele funcione conforme esperado em situações de alto tráfego.
Entrega:

O código-fonte completo da implementação.
Documentação explicando como o rate limiter funciona e como ele pode ser configurado.
Testes automatizados demonstrando a eficácia e a robustez do rate limiter.
Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.
O servidor web deve responder na porta 8080.

# Pre-requisitos

- Go 1.24 ou superior
- Redis
- Docker e Docker Compose

# Como executar
1. Clone o repositório:
   ```bash
   git clone https://github.com/fhsmendes/rate-limiter.git
   ```
2. Acesse o diretório do projeto:
   ```bash
   cd rate-limiter
   ```
3. Crie um arquivo `.env` com as variáveis de ambiente necessárias:
   
| Nome | Valor de exemplo | Descrição |
|------|------------------|-----------|
| PORT | 8080 | Porta onde o servidor web será executado |
| IP_RATE_LIMIT | 10 | Número máximo de requisições por segundo por IP |
| IP_RATE_LIMIT_DURATION_SECONDS | 30 | Duração em segundos para o limite de requisições por IP |
| IP_BLOCK_DURATION_SECONDS | 45 | Duração em segundos do bloqueio quando o limite é excedido |
| JWT_SECRET | "SECRET" | Chave secreta para validação de tokens JWT |
| JWT_EXPIRES_IN | 120 | Tempo de expiração do token em segundos |
| REDIS_ENABLED | true | Habilita ou desabilita o uso do Redis |
| REDIS_HOST | "localhost:6379" | Endereço do servidor Redis |
| REDIS_PASSWORD | "" | Senha do Redis (deixe vazio se não houver senha) |
| REDIS_DB | 0 | Número do banco de dados Redis a ser utilizado |

- estrutura do arquivo `.env` deve ser semelhante a esta:

    Exemplo de conteúdo do arquivo `.env`:

    ```env
    PORT=8080
    IP_RATE_LIMIT=10
    IP_RATE_LIMIT_DURATION_SECONDS=30
    IP_BLOCK_DURATION_SECONDS=45
    JWT_SECRET="SECRET"
    JWT_EXPIRES_IN=120
    REDIS_ENABLED=true
    REDIS_HOST="localhost:6379"
    REDIS_PASSWORD=""
    REDIS_DB=0
    ```
 - você pode criar um token JWT com as seguintes claims:

    ```json
    {
      "rate_limit": 10,
      "rate_limit_duration": 60,
      "block_duration": 300
    }
    ```

- Exemplos de uso

    #### Fazendo requisições com curl

    ##### Requisição sem token (limitada por IP):
    
    ```bash
    curl -X GET http://localhost:8080/token \
    -H "Content-Type: application/json" \
    -d '{
        "rate_limit": 10,
        "rate_limit_duration": 60,
        "block_duration": 300
    }'
    ```

4. Inicie o Redis usando Docker Compose:
   ```bash
   docker-compose up -d
   ```
5. Atualize as dependências:
   ```bash
   go mod tidy
   ```

6. Para subir o servidor web, execute:

    >porta padrão é a 8080       
    
    ```bash
   go run main.go
   ```

7. Testes automatizados:
   - Para executar os testes, utilize o comando:
   >Para agilizar os testes os limites foram reduzidos

   ```bash
   go test -v ./main_test.go
   ```