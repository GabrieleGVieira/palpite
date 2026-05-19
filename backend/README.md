# PalpitAI Backend

Backend em Go do PalpitAI. Ele expoe a API HTTP, valida usuarios autenticados pelo Supabase Auth, persiste dados no Supabase/Postgres, emite eventos realtime por WebSocket e sincroniza placares da Copa do Mundo via football-data.org.

## Requisitos

- Go 1.24+
- Banco Supabase/Postgres
- Redis Upstash
- URL e chave publica do Supabase
- Token do football-data.org

## Configuracao

```bash
cp .env.example .env
```

Variaveis:

```bash
APP_ENV=development
PORT=3000
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres
REDIS_URL=redis://default:token@host.upstash.io:6379
SUPABASE_URL=https://project.supabase.co
SUPABASE_KEY=cole_a_chave_publica_aqui
FOOTBALL_DATA_API_BASE_URL=https://api.football-data.org/v4
FOOTBALL_DATA_COMPETITION_CODE=WC
FOOTBALL_DATA_SEASON=2026
FOOTBALL_DATA_TOKEN=cole_o_token_aqui
```

O backend adiciona `sslmode=require` automaticamente quando `DATABASE_URL` nao informa `sslmode`.
O cliente Redis usa TLS automaticamente para conectar no Upstash via `REDIS_URL`.

## Como rodar

API HTTP/WebSocket:

```bash
make run
```

Worker de sincronizacao de jogos:

```bash
go run ./cmd/matchsync
```

Seed inicial:

```bash
make seed
```

## Arquitetura

```text
backend/
├── cmd/
│   ├── api/          # entrada da API HTTP/WebSocket
│   ├── matchsync/    # worker separado para sincronizar partidas
│   └── seed/         # comando de seed
├── docs/
│   └── realtime-match-sync.md
└── internal/
    ├── apperrors/    # erros de aplicacao
    ├── config/       # leitura de env/config
    ├── controller/   # handlers HTTP/WebSocket
    ├── database/     # conexao, migrations e seed
    ├── domain/       # regras puras de dominio
    ├── dto/          # contratos externos/HTTP e mappers
    ├── matchsync/    # client, scheduler, syncer e publicacao de eventos
    ├── realtime/     # hub WebSocket e broadcast
    ├── repositories/ # queries SQL por tabela/contexto
    ├── route/        # roteamento HTTP
    ├── usecase/      # casos de uso da aplicacao
    └── utils/        # helpers compartilhados
```

Separacao principal:

- `controller`: parseia request, chama usecase e escreve response.
- `usecase`: orquestra regras de negocio e transacoes.
- `repositories`: concentra SQL e acesso ao banco.
- `domain`: mantem regras puras, como calculo de pontos e normalizacao de status.
- `dto`: define payloads HTTP e modelos da API football-data.org.
- `matchsync`: sincroniza jogos, detecta alteracoes e publica eventos.
- `realtime`: gerencia conexoes WebSocket e envia eventos para grupos.

## Rotas

```text
GET /health
GET /ws
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

Todas as rotas de usuario exigem `Authorization: Bearer <access_token>` do Supabase Auth.

## Grupos

### Criar grupo

`POST /api/v1/groups`

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

A rota cria o grupo, gera `invite_code` e adiciona o usuario autenticado em `group_members` como `owner`.

### Listar meus grupos

`GET /api/v1/groups`

Retorna os grupos em que o usuario participa, incluindo informacoes usadas pela Home, como pontuacao e pendencias de aprovacao quando o usuario e owner.

### Entrar em grupo

`POST /api/v1/groups/join`

```json
{
  "invite_code": "ABCD1234"
}
```

Se o grupo for publico, o usuario entra como membro ativo. Se o grupo for privado, a entrada fica pendente ate o owner aprovar.

## Jogos, palpites e ranking

`GET /api/v1/groups/{groupID}/matches` retorna os jogos do grupo e o palpite do usuario autenticado quando existir.

`PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction` salva ou edita o palpite antes do inicio do jogo.

`PUT /api/v1/matches/{matchID}/result` registra ou altera um placar manualmente, recalcula pontos e emite eventos realtime quando houver mudanca.

`GET /api/v1/me/score` retorna a pontuacao geral do usuario autenticado.

`GET /api/v1/groups/{groupID}/ranking` retorna posicao, usuario e pontuacao total dos participantes ativos do grupo.

Pontuacao atual:

- Placar exato: 10 pontos
- Acertou vencedor ou empate: 5 pontos
- Errou tudo: 0 pontos

## Sincronizacao de partidas

O comando `cmd/matchsync` consome:

```text
GET https://api.football-data.org/v4/competitions/WC/matches
```

E usa polling adaptativo:

- Jogos ao vivo: a cada 30s
- Jogos do dia: a cada 5min
- Proximos jogos: a cada 1h
- Limite respeitado: 10 requests/min, com intervalo minimo entre chamadas

O syncer compara resposta externa com o estado atual do banco. Ele atualiza o banco e publica eventos apenas quando existe alteracao real de status, placar, gols ou resultado final.

Mais detalhes: [`docs/realtime-match-sync.md`](docs/realtime-match-sync.md).

## Realtime

`GET /ws?token=<access_token>&group_id=<groupID>` abre uma conexao WebSocket autenticada.

Eventos emitidos:

```json
{
  "type": "match.updated",
  "group_id": "uuid",
  "payload": {
    "match_id": "uuid"
  }
}
```

```json
{
  "type": "ranking.updated",
  "group_id": "uuid",
  "payload": {
    "group_id": "uuid"
  }
}
```

```json
{
  "type": "match.finished",
  "group_id": "uuid",
  "payload": {
    "message": "Brasil 2x1 Croacia - resultado final lancado"
  }
}
```

O frontend usa esses eventos para invalidar dados de jogos/ranking e mostrar notificacoes temporarias.

## Banco

As migrations sao aplicadas na inicializacao da API. As principais tabelas sao:

- `groups`
- `group_members`
- `world_cup_matches`
- `predictions`
- `match_events`

Regras importantes:

- Owner tambem e membro do grupo.
- Grupo privado cria solicitacao pendente antes de virar membro ativo.
- Palpite so pode ser editado antes do inicio da partida.
- Ranking e derivado dos palpites pontuados por grupo.

## Qualidade

```bash
make fmt
make vet
make test
```
