# Realtime Match Sync

Arquitetura MVP para sincronizar jogos da Copa via `football-data.org`, atualizar ranking e emitir eventos realtime com baixo consumo de requests.

## Objetivo

- Buscar jogos em `GET /v4/competitions/WC/matches`.
- Evitar polling desnecessario.
- Não reprocessar jogos sem mudanca.
- Atualizar ranking somente quando pontos mudarem.
- Emitir eventos WebSocket somente quando houver diff.
- Manter a arquitetura pronta para Redis/pubsub sem exigir Redis no MVP.

## Estrutura Go

```text
backend/internal/
├── config/                 # env vars
├── database/               # migrations e conexao Postgres
├── httpapi/                # REST API
├── matchsync/              # football-data client + scheduler + diff
│   └── syncer.go
├── realtime/               # hub WebSocket
└── ranking/                # futuro calculo/materializacao de ranking
```

No MVP, `matchsync` publica eventos no `realtime.Hub`, que entrega para clientes WebSocket conectados em `/ws`.

## Fluxo Realtime

1. Scheduler dispara sync `LIVE`, `TODAY` ou `UPCOMING`.
2. Cliente chama `football-data.org` respeitando 10 req/min.
3. Backend normaliza status e placar.
4. Backend compara API x DB.
5. Se não mudou: não escreve no DB e não emite evento.
6. Se mudou: atualiza `world_cup_matches`.
7. Se houver gols novos: insere em `match_events` e emite `match.goal`.
8. Se placar `live` ou `finished`: recalcula pontos alterados.
9. Se pontos mudaram: emite `ranking.updated`.
10. App escuta eventos WebSocket e recarrega ranking/jogos.

## Polling Inteligente

- `LIVE`: a cada 30s, mas apenas se houver jogo `live` no DB ou jogo entre `now - 3h` e `now + 30min`.
- `TODAY`: a cada 5min, busca jogos do dia sem depender do DB.
- `UPCOMING`: a cada 1h, busca proximos 30 dias para manter agenda atualizada.
- Rate limit global: 1 request a cada 6s por processo.

Essa estrategia fica abaixo de 10 req/min e evita gastar chamada live fora de janela de jogo.

## Detecção de Mudanças

O `upsert` usa `IS DISTINCT FROM` nos campos relevantes:

- `external_id`
- `stage`
- `status`
- `home_score`
- `away_score`

Se nada mudou, `RowsAffected = 0`; logo, não recalcula ranking nem publica websocket.

Para eventos:

- `match_events.external_key` e unico.
- Gols usam chave deterministica: `matchID:goal:minute:teamID:scorerID:homeScore:awayScore`.
- Eventos repetidos da API sao ignorados com `ON CONFLICT DO NOTHING`.

## Banco

Tabelas principais:

- `world_cup_matches`: agenda, status, placar, sync.
- `predictions`: palpites e pontos.
- `group_members`: usuarios ativos no grupo.
- `groups`: escopo do grupo.
- `match_events`: gols e futuros eventos de jogo.

Campos importantes em `world_cup_matches`:

- `external_id`
- `status`: `scheduled`, `live`, `finished`, `postponed`, `cancelled`
- `home_score`, `away_score`
- `finished_at`
- `last_synced_at`

Indices:

- `(status, kickoff_at)` para janela live.
- `external_id where external_id is not null`.
- `match_events(match_id, event_type, minute)`.

## Eventos Websocket

Rooms sugeridas:

- `group:{groupID}`
- `match:{matchID}`
- `rankings`
- `matches`

Eventos:

```json
{
  "name": "match.updated",
  "room": "matches",
  "payload": {
    "external_id": "329400",
    "home_team": "Brasil",
    "away_team": "Croacia",
    "home_score": 2,
    "away_score": 1,
    "status": "live",
    "previous_status": "scheduled"
  }
}
```

```json
{
  "name": "match.finished",
  "room": "matches",
  "payload": {
    "home_team": "Brasil",
    "away_team": "Croacia",
    "home_score": 2,
    "away_score": 1,
    "status": "finished"
  }
}
```

```json
{
  "name": "match.goal",
  "room": "match:uuid",
  "payload": {
    "match_id": "uuid",
    "minute": 84,
    "team_name": "Brasil",
    "player_name": "Youssoufa Moukoko",
    "home_score": 2,
    "away_score": 1
  }
}
```

```json
{
  "name": "ranking.updated",
  "room": "rankings",
  "payload": {
    "home_team": "Brasil",
    "away_team": "Croacia",
    "home_score": 2,
    "away_score": 1
  }
}
```

No MVP, o app recebe `ranking.updated` e chama `GET /api/v1/groups/{groupID}/ranking` automaticamente.

## Ranking Eficiente

MVP:

- Recalcular pontos apenas dos palpites do jogo alterado.
- Usar `p.points IS DISTINCT FROM new_points` para atualizar somente quem mudou.
- Ranking continua calculado por query agregada.

Proxima etapa:

- Criar `group_rankings(group_id, user_id, total_points, position, updated_at)`.
- Atualizar apenas grupos afetados pelo `match_id`.
- Publicar evento por `group:{groupID}`.

Query de grupos afetados:

```sql
select distinct group_id
from predictions
where match_id = $1;
```

## Cache

MVP em memoria:

- Rate limiter por processo.
- Mutex para impedir sync concorrente.
- DB como fonte da verdade.

Futuro com Redis:

- `SETNX lock:match-sync live EX 25` para lock distribuido.
- Cache da ultima resposta por filtro por 20s-60s.
- Pub/sub de eventos: `match.updated`, `ranking.updated`.
- Deduplicacao de eventos com `SETNX event:{external_key}`.

## Race Conditions

Medidas atuais:

- `sync.Mutex.TryLock()` evita dois syncs no mesmo processo.
- `rateMu` serializa requests ao provedor.
- `ON CONFLICT` torna upsert idempotente.
- `match_events.external_key unique` evita evento duplicado.
- `p.points IS DISTINCT FROM` evita reescrita e evento de ranking sem mudanca.

Futuro multi-instancia:

- Trocar mutex local por Redis lock ou Postgres advisory lock.
- Publicar eventos via Redis pub/sub.
- Apenas uma instancia deve executar schedulers.

## Workers

MVP:

- Uma goroutine para `matchsync.Run`.
- Tres tickers no mesmo worker:
  - `LIVE`: 30s
  - `TODAY`: 5min
  - `UPCOMING`: 1h

Futuro:

- `scheduler` publica jobs.
- `worker pool` consome jobs com lock distribuido.
- `publisher` envia eventos WebSocket.

## Clean Architecture

Separar interfaces:

- `Provider`: busca dados externos.
- `MatchRepository`: upsert/diff.
- `PredictionScorer`: recalcula pontos.
- `EventRepository`: persiste eventos de jogo.
- `Publisher`: WebSocket/Redis/log.
- `Scheduler`: decide quando rodar.

No MVP, `syncer.go` concentra isso para manter velocidade. Quando crescer, extraia por responsabilidade sem mudar contrato externo.
