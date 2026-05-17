# PalpitAI Backend

API inicial em Go para o PalpitAI.

## Requisitos

- Go 1.24+

## Como rodar

```bash
cd backend
cp .env.example .env
make run
```

A API inicia em `http://localhost:3000`.

Configure `DATABASE_URL` no `.env` com a connection string PostgreSQL do Supabase. O backend adiciona `sslmode=require` automaticamente quando a URL nao informa `sslmode`.

Configure tambem `SUPABASE_URL` e `SUPABASE_KEY` para validar o token recebido do Supabase Auth.

## Rotas iniciais

```text
GET /health
GET /api/v1/status
GET /api/v1/me/score
GET /api/v1/groups
POST /api/v1/groups
PUT /api/v1/groups/{groupID}
POST /api/v1/groups/join
GET /api/v1/groups/{groupID}/matches
GET /api/v1/groups/{groupID}/ranking
PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction
PUT /api/v1/matches/{matchID}/result
```

As respostas incluem o status da conexao com o banco:

```json
{
  "database": "ok",
  "status": "ok"
}
```

### Criar grupo

`POST /api/v1/groups` exige `Authorization: Bearer <access_token>` do Supabase Auth.

Payload:

```json
{
  "name": "Familia na Copa",
  "description": "Bolao da familia",
  "match_scope": "selected",
  "selected_teams": ["Brasil", "Argentina"],
  "participant_limit": null,
  "has_unlimited_participants": true,
  "is_private": true
}
```

A rota cria o grupo, gera `invite_code` e adiciona o usuario autenticado em `group_members` com role `owner`.

### Meus grupos

`GET /api/v1/groups` exige `Authorization: Bearer <access_token>` do Supabase Auth e retorna os grupos em que o usuario autenticado participa.

### Entrar no grupo

`POST /api/v1/groups/join` exige `Authorization: Bearer <access_token>` do Supabase Auth.

Payload:

```json
{
  "invite_code": "ABCD1234"
}
```

A rota adiciona o usuario autenticado em `group_members` como `member`, respeitando o limite de participantes quando existir.

### Jogos e palpites

`GET /api/v1/groups/{groupID}/matches` retorna os jogos do grupo e o palpite do usuario autenticado quando existir.

`PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction` salva ou edita o palpite antes do inicio do jogo.

### Resultado e pontuacao

`PUT /api/v1/matches/{matchID}/result` registra o placar final e calcula os pontos dos palpites da partida:

- placar exato: 10 pontos
- acertou vencedor ou empate: 5 pontos
- errou tudo: 0 pontos

`GET /api/v1/me/score` retorna a pontuacao geral do usuario autenticado, somando os palpites pontuados em todos os grupos ativos.

`GET /api/v1/groups/{groupID}/ranking` retorna o ranking do grupo com posicao, usuario e pontuacao total de cada participante ativo.

## Comandos

```bash
make run
make test
make fmt
make vet
```
