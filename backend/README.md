# Palpite! Backend

API HTTP/WebSocket em Go do Palpite! A API autentica Palpiteiros via Supabase Auth, persiste dados no PostgreSQL, emite eventos em tempo real por WebSocket, sincroniza placares da Copa do Mundo via football-data.org e sustenta os módulos sociais, Palpicoins, desafios e PalpitAI.

## O que é

O backend é o núcleo do sistema: recebe palpites, calcula pontuação, serve rankings, administra grupos e coordena atualizações em tempo real para todos os clientes conectados. Um worker separado faz polling na API da football-data.org e propaga mudanças de placar e status de partidas via WebSocket. O backend também integra com o pipeline da PalpitAI via banco de dados, gera explicações com a Gemini API e mantém funcionalidades sociais como amizades, perfis públicos, feed de grupo, pagamentos, carteira de Palpicoins e desafios.

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
BETA_ANDROID_PLAY_STORE_URL=https://play.google.com/apps/testing/com.gabrielevieira.palpite
BETA_APPROVAL_BASE_URL=https://seudominio.com
BETA_APPROVAL_SECRET=segredo_longo_para_assinar_links
BETA_SIGNUP_NOTIFICATION_EMAIL=seu-email
EMAIL_PROVIDER=brevo
EMAIL_FROM_NAME=Palpite!
EMAIL_FROM_ADDRESS=email-configurado-no-brevo@example.com
BREVO_SMTP_HOST=smtp-relay.brevo.com
BREVO_SMTP_PORT=587
BREVO_SMTP_USER=login-smtp-do-brevo
BREVO_SMTP_PASSWORD=chave-smtp-do-brevo
AI_EXPLANATION_BATCH_SIZE=2
AI_EXPLANATION_MIN_BATCH_SIZE=1
AI_EXPLANATION_RETRY_MISSING=true
AI_EXPLANATION_MAX_MISSING_RETRIES=2
AI_EXPLANATION_SEED_DAYS=90
AI_EXPLANATION_REFRESH_DAYS=7
AI_EXPLANATION_MAX_AGE_HOURS=24
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
make matchsync
```

**Worker de explicações da PalpitAI:**

```bash
make explanations MODE=seed LIMIT=50
```

`MODE=seed` usa uma janela futura a partir da data atual. `MODE=refresh` reprocessa previsões recentes conforme `AI_EXPLANATION_REFRESH_DAYS`.

**Migrations:**

```bash
make migrate
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
    ├── social/       # eventos sociais do feed
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
- `explanations` — gera e persiste explicações estruturadas da PalpitAI
- `goalpredictions`, `metrics`, `predictions` — leitura dos artefatos do pipeline ML
- `repositories/*payments`, `friends`, `palpicoins`, `challenges`, `feed` — persistência dos módulos sociais e econômicos virtuais

## Rotas

```text
GET    /health
GET    /ws
GET    /api/v1/status
DELETE /api/v1/me
GET    /api/v1/me/profile
PATCH  /api/v1/me/profile
GET    /api/v1/me/score
GET    /api/v1/me/wallet
GET    /api/v1/me/wallet/transactions
GET    /api/v1/rankings/palpicoins
POST   /api/v1/friends/request
POST   /api/v1/friends/{id}/accept
POST   /api/v1/friends/{id}/decline
DELETE /api/v1/friends/{id}
GET    /api/v1/friends
GET    /api/v1/friends/requests
GET    /api/v1/users/search
GET    /api/v1/users/{id}/profile
POST   /api/v1/challenges
POST   /api/v1/challenges/{id}/accept
POST   /api/v1/challenges/{id}/decline
POST   /api/v1/challenges/{id}/cancel
GET    /api/v1/challenges
GET    /api/v1/challenges/{id}
GET    /api/v1/matches/{matchID}/prediction
GET    /api/v1/groups
POST   /api/v1/groups
PUT    /api/v1/groups/{groupID}
POST   /api/v1/groups/join
GET    /api/v1/groups/{groupID}/join-requests
POST   /api/v1/groups/{groupID}/join-requests/{userID}/approve
GET    /api/v1/groups/{groupID}/members
GET    /api/v1/groups/{groupID}/members/{userID}
GET    /api/v1/groups/{groupID}/feed
POST   /api/v1/groups/{groupID}/feed/{eventID}/reaction
DELETE /api/v1/groups/{groupID}/feed/{eventID}/reaction
POST   /api/v1/groups/{groupID}/members/{userID}/transfer-ownership
DELETE /api/v1/groups/{groupID}/members/{userID}
GET    /api/v1/groups/{groupID}/payments
GET    /api/v1/groups/{groupID}/payments/summary
PATCH  /api/v1/groups/{groupID}/payments/{userID}
DELETE /api/v1/groups/{groupID}/membership
GET    /api/v1/groups/{groupID}/matches
GET    /api/v1/groups/{groupID}/ranking
PUT    /api/v1/groups/{groupID}/matches/{matchID}/prediction
PUT    /api/v1/matches/{matchID}/result
POST   /api/beta/android
PATCH  /admin/beta-testers/{id}/approve
```

Todas as rotas de usuário exigem `Authorization: Bearer <access_token>` do Supabase Auth.
`POST /api/beta/android` é público e tem rate limit simples por IP. A rota `/admin/beta-testers/{id}/approve` exige Bearer token e usa o usuário autenticado como responsável pela aprovação.

## Beta Android

O fluxo da landing é:

```text
Landing form -> POST /api/beta/android -> beta_testers_android(status=pending_approval)
```

O signup público valida consentimento/e-mail, salva ou atualiza o registro, envia um alerta via Brevo SMTP para `BETA_SIGNUP_NOTIFICATION_EMAIL` e retorna sucesso para a landing. Se `BETA_APPROVAL_BASE_URL` e `BETA_APPROVAL_SECRET` estiverem configurados, esse alerta inclui um botão de confirmação assinado para aprovar o tester depois que o e-mail for adicionado no Play Console. Se o envio de e-mail falhar ou `EMAIL_PROVIDER` estiver vazio em ambiente local/teste, o cadastro continua funcionando e a falha ou desativação fica apenas no log.

`BETA_ANDROID_PLAY_STORE_URL` é usado no e-mail de aprovação enviado ao tester. A landing não redireciona automaticamente após o cadastro.

A aprovação manual usa `PATCH /admin/beta-testers/{id}/approve`, altera o status para `approved`, preenche `approved_at`/`approved_by` e envia o e-mail de liberação ao tester. Se o envio falhar, a aprovação permanece salva.

O link assinado do e-mail administrativo abre a landing em `/admin/beta-testers/{id}/approve/confirm` com uma página de revisão. A landing chama a API do backend para carregar os dados e confirmar a aprovação, reduzindo o risco de aprovação por scanners de e-mail. Configure `BETA_APPROVAL_BASE_URL` com a URL pública da landing.

O adapter de Google Groups permanece isolado em `internal/google` para uso futuro, mas não faz parte do signup público da landing.

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
| `202605260005_add_prediction_explanation_refresh_fields` | Campos de refresh, erro e auditoria em `prediction_explanations` |
| `202605290001_create_group_payments` | Campos de bolão pago em `groups` e status de pagamento em `group_members` |
| `202605290002_add_group_member_avatar_url` | Snapshot de avatar em membros de grupo |
| `202605290003_create_group_feed_events` | `group_feed_events`, reações e índices do feed |
| `202606010001_allow_multiple_feed_reactions` | Múltiplas reações por evento de feed |
| `202606010002_create_friendships` | `friendships` |
| `202606010003_create_user_social_settings` | Configurações de perfil público |
| `202606010004_create_palpicoins_and_challenges` | Carteira, transações, ranking e desafios com Palpicoins |

## Qualidade

```bash
make fmt
make vet
make test
```
