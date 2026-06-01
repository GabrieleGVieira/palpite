# Palpite! Backend

API HTTP/WebSocket em Go do Palpite! A API autentica Palpiteiros via Supabase Auth, persiste dados no PostgreSQL, emite eventos em tempo real por WebSocket e sincroniza placares da Copa do Mundo via football-data.org.

## O que é

O backend é o núcleo do sistema: recebe palpites, calcula pontuação, serve o ranking e coordena atualizações em tempo real para todos os clientes conectados. Um worker separado faz polling na API da football-data.org e propaga mudanças de placar e status de partidas via WebSocket. O backend também integra com o pipeline da PalpitAI via banco de dados e gera explicações com a Gemini API.

## Tecnologias

- **Go 1.24+** — linguagem principal
- **net/http** — servidor HTTP nativo
- **gorilla/websocket** — conexões WebSocket
- **jackc/pgx** — driver PostgreSQL
- **go-redis** — cliente Redis
- **PostgreSQL** (Supabase) — banco de dados principal
- **Redis** (Upstash) — cache e pub/sub
- **Gemini API** — geração de explicações de previsões
- **football-data.org** — fonte de placares e resultados

## Fontes de dados

| Fonte                  | Uso                                                    |
| ---------------------- | ------------------------------------------------------ |
| football-data.org API  | Placares, status e resultados das partidas em tempo real |
| PostgreSQL (Supabase)  | Grupos, membros, palpites, ranking e análises de ML   |
| Redis (Upstash)        | Cache de sessões e estado do hub WebSocket             |
| Gemini API            | Explicações textuais das previsões geradas pelo ML     |

## Configuração

```bash
cp .env.example .env
```

```env
APP_ENV=development
PORT=3000
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres
REDIS_URL=redis://default:token@host.upstash.io:6379
SUPABASE_URL=https://project.supabase.co
SUPABASE_KEY=chave_publica
SUPABASE_SERVICE_ROLE_KEY=chave_service_role_para_excluir_auth_opcional
FOOTBALL_DATA_API_BASE_URL=https://api.football-data.org/v4
FOOTBALL_DATA_COMPETITION_CODE=WC
FOOTBALL_DATA_SEASON=2026
FOOTBALL_DATA_TOKEN=token
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.5-flash
GEMINI_RATE_LIMIT_COOLDOWN_SECONDS=1800
GEMINI_RATE_LIMIT_MAX_WAITS=1
GEMINI_REQUEST_DELAY_SECONDS=15
GEMINI_TIMEOUT_SECONDS=30
```

Se a conexão direta ao Supabase falhar por IPv6/TLS, use a URL do Transaction Pooler:

```env
DATABASE_URL=postgresql://postgres.project-ref:password@aws-0-region.pooler.supabase.com:6543/postgres
```

## Como rodar

**API HTTP/WebSocket:**

```bash
make run
```

**Worker de sincronização de partidas:**

```bash
go run ./cmd/matchsync
```

**Migrations:**

```bash
go run ./cmd/migrate
```

**Seed inicial:**

```bash
make seed
```

## Arquitetura

```text
backend/
├── cmd/
│   ├── api/          # entrada da API HTTP/WebSocket
│   ├── matchsync/    # worker de sincronização de partidas
│   ├── migrate/      # aplicação de migrations SQL
│   └── seed/         # carga inicial de dados
└── internal/
    ├── ai/           # cliente Gemini e geração de explicações
    ├── apperrors/    # erros de aplicação tipados
    ├── cache/        # abstração de cache Redis
    ├── config/       # leitura de variáveis de ambiente
    ├── controller/   # handlers HTTP/WebSocket
    ├── database/     # conexão e setup do banco
    ├── domain/       # regras puras de domínio (pontuação, status)
    ├── dto/          # contratos HTTP e mappers
    ├── explanations/ # serviço de explicações de previsões
    ├── matchsync/    # client, scheduler, syncer e publicação de eventos
    ├── ml/           # integração com previsões do pipeline ML
    ├── predictions/  # leitura de previsões salvas pelo ML service
    ├── realtime/     # hub WebSocket e broadcast por grupo
    ├── repositories/ # queries SQL por contexto
    ├── route/        # roteamento HTTP
    ├── usecase/      # casos de uso e orquestração de negócio
    └── utils/        # helpers compartilhados
```

**Separação de responsabilidades:**

- `controller` — parseia request, chama usecase, escreve response
- `usecase` — orquestra regras de negócio e transações
- `repositories` — concentra SQL e acesso ao banco
- `domain` — mantém regras puras como cálculo de pontos e normalização de status
- `matchsync` — sincroniza partidas, detecta alterações e publica eventos
- `realtime` — gerencia conexões WebSocket e envia eventos por grupo

## Rotas

```text
GET  /health
GET  /ws
GET  /api/v1/status
DELETE /api/v1/me
GET  /api/v1/me/score
GET  /api/v1/groups
POST /api/v1/groups
PUT  /api/v1/groups/{groupID}
POST /api/v1/groups/join
GET  /api/v1/groups/{groupID}/matches
GET  /api/v1/groups/{groupID}/ranking
PUT  /api/v1/groups/{groupID}/matches/{matchID}/prediction
PUT  /api/v1/matches/{matchID}/result
```

Todas as rotas de usuário exigem `Authorization: Bearer <access_token>` do Supabase Auth.

## Realtime (WebSocket)

```
GET /ws?token=<access_token>&group_id=<groupID>
```

Eventos emitidos:

```json
{ "type": "match.updated",   "group_id": "uuid", "payload": { "match_id": "uuid" } }
{ "type": "ranking.updated", "group_id": "uuid", "payload": { "group_id": "uuid" } }
{ "type": "match.finished",  "group_id": "uuid", "payload": { "message": "Brasil 2x1 Croácia - resultado final lançado" } }
```

## Sincronização de partidas

O worker `cmd/matchsync` faz polling em `GET https://api.football-data.org/v4/competitions/WC/matches` com frequência adaptativa:

| Situação              | Intervalo |
| --------------------- | --------- |
| Jogos ao vivo         | 30s        |
| Jogos do dia          | 5min       |
| Próximos jogos        | 1h         |

Atualiza o banco e publica eventos WebSocket somente quando há mudança real de status, placar ou resultado.

## Banco de dados

O schema é gerenciado por dois mecanismos complementares:

- **`cmd/migrate`** — aplica os arquivos `.sql` em `migrations/` em ordem lexicográfica. Deve ser executado antes de iniciar a API em um banco novo.
- **`database/migrations.go`** — executado automaticamente no startup da API via `database.Migrate()`. Cria as tabelas core da app com `IF NOT EXISTS`, servindo como fallback idempotente.

Em um banco novo, rode `cmd/migrate` primeiro — ele cobre todas as tabelas.

Arquivos de migration:

| Arquivo | Tabelas |
| --- | --- |
| `202605230001_create_core_app_tables` | `groups`, `group_members`, `world_cup_matches`, `match_events`, `predictions` |
| `202605240001_create_data_pipeline_tables` | `teams`, `team_aliases` |
| `202605250001_create_metrics_tables` | `team_metrics`, `team_metric_snapshots`, `match_features` |
| `202605260001_create_ml_prediction_tables` | `ml_models`, `prediction_runs`, `historical_matches`, `match_predictions` |
| `202605260002_create_goal_prediction_tables` | `goal_models`, `match_goal_predictions`, `match_score_probabilities` |
| `202605260003_add_score_result_calibration` | Colunas de calibração em `match_goal_predictions` |
| `202605260004_create_prediction_explanations` | `prediction_explanations` |

## Qualidade

```bash
make fmt
make vet
make test
```
