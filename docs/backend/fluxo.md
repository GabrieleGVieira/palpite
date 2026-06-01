# Backend — Fluxo do Sistema

Documentação dos fluxos principais do backend Go do Palpite!: ciclo de vida de requisições HTTP, WebSocket, sincronização de partidas e geração de explicações da PalpitAI.

---

## Visão geral da arquitetura

```mermaid
graph TD
    CLIENT["App Mobile (Frontend)"]
    SUPABASE["Supabase Auth"]
    API["API Go (cmd/api)"]
    WS["Hub WebSocket"]
    DB["PostgreSQL (Supabase)"]
    REDIS["Redis (Upstash)"]
    FOOTBALL["football-data.org"]
    GEMINI["Gemini API"]
    MATCHSYNC["Worker MatchSync (cmd/matchsync)"]
    WORKER["Worker Explanations"]
    MLDB["Tabelas ML no banco"]

    CLIENT -- "HTTP + Bearer token" --> API
    CLIENT -- "ws://" --> WS
    API -- "valida JWT" --> SUPABASE
    API -- "lê/escreve" --> DB
    API -- "cache" --> REDIS
    MATCHSYNC -- "polling GET /matches" --> FOOTBALL
    MATCHSYNC -- "atualiza" --> DB
    MATCHSYNC -- "publica eventos" --> WS
    WORKER -- "lê previsões" --> MLDB
    WORKER -- "gera explicações" --> GEMINI
    WORKER -- "salva explicações" --> DB
    WS -- "broadcast por grupo" --> CLIENT
```

---

## 1. Ciclo de vida de uma requisição HTTP

Toda requisição protegida passa por autenticação, depois desce pela stack controller → usecase → repository.

```mermaid
sequenceDiagram
    participant App as App Mobile
    participant Controller
    participant Auth as Supabase Auth
    participant Usecase
    participant Repository
    participant DB as PostgreSQL

    App->>Controller: PUT /api/v1/groups/{gid}/matches/{mid}/prediction
    Note right of App: Header: Authorization Bearer <token>

    Controller->>Auth: GET {supabaseURL}/auth/v1/user
    Note right of Controller: Header: apikey + Bearer token
    Auth-->>Controller: { id, email, user_metadata }

    Controller->>Controller: parseia body JSON<br/>{ home_score, away_score }

    Controller->>Usecase: SavePrediction(ctx, userID, groupID, matchID, req)

    Usecase->>Repository: EnsureActiveGroupMember(userID, groupID)
    Repository->>DB: SELECT FROM group_members WHERE ...
    DB-->>Repository: membro ativo ✓

    Usecase->>Repository: MatchKickoffForGroup(groupID, matchID)
    Repository->>DB: SELECT kickoff_at FROM world_cup_matches
    DB-->>Repository: kickoff_at

    Usecase->>Usecase: valida time.Now() < kickoff_at

    Usecase->>Repository: UpsertPrediction(userID, groupID, matchID, scores)
    Repository->>DB: INSERT ... ON CONFLICT DO UPDATE RETURNING *
    DB-->>Repository: { match_id, home_score, away_score, points, updated_at }

    Repository-->>Usecase: Prediction
    Usecase-->>Controller: Prediction
    Controller-->>App: 200 OK { match_id, home_score, away_score, points }
```

**Camadas e responsabilidades:**

| Camada | Responsabilidade |
| --- | --- |
| `controller` | Parse de request, validação de entrada, escrita de response |
| `usecase` | Orquestra regras de negócio, chama múltiplos repositories |
| `repository` | SQL isolado por contexto, sem lógica de negócio |
| `domain` | Regras puras (cálculo de pontos, normalização de status) |

---

## 2. Autenticação via Supabase

A validação de JWT ocorre em toda requisição protegida, sem estado local no backend.

```mermaid
flowchart LR
    A["Requisição chega"] --> B{"Authorization\nheader presente?"}
    B -- Não --> ERR1["401 Unauthorized"]
    B -- Sim --> C["Extrai Bearer token"]
    C --> D["GET {supabaseURL}/auth/v1/user\nApikey + Bearer token"]
    D --> E{"Status 200?"}
    E -- Não --> ERR2["401 Unauthorized"]
    E -- Sim --> F{"user.ID\nnão vazio?"}
    F -- Não --> ERR3["401 Unauthorized"]
    F -- Sim --> G["Requisição prossegue\ncom userID extraído"]
```

---

## 3. WebSocket — conexão e ciclo de vida

O WebSocket permite que o backend empurre eventos em tempo real para o frontend sem polling.

```mermaid
sequenceDiagram
    participant App as App Mobile
    participant Handler as RealtimeHandler
    participant Hub as WebSocket Hub
    participant Publisher

    App->>Handler: GET /ws?token=<jwt>&group_id=<uuid>
    Handler->>Handler: valida token via Supabase
    Handler->>Handler: valida membership no grupo (opcional)
    Handler->>Hub: ServeWS(conn, userID, rooms)

    Hub->>Hub: Upgrade HTTP → WebSocket
    Hub->>Hub: Cria Client { conn, userID, rooms }
    Hub->>Hub: registra client no hub.clients

    par writePump
        Hub-->>App: escreve eventos (json.Marshal(event))
        Hub-->>App: ping frame a cada 45s
    and readPump
        App->>Hub: pong frame (reset deadline para 60s)
    end

    Note over App,Hub: Rooms subscritas:<br/>• matches (todos os jogos)<br/>• rankings (ranking global)<br/>• user:<userID><br/>• group:<groupID> (se informado)

    Publisher->>Hub: Publish(event)
    Hub->>Hub: itera clients, filtra por room
    Hub->>App: envia event.JSON no send channel
```

### Estrutura de um evento

```json
{
  "name": "match.updated",
  "payload": {
    "match_id": "uuid",
    "home_score": 2,
    "away_score": 1,
    "status": "live"
  },
  "room": "matches"
}
```

| Evento | Room | Trigger |
| --- | --- | --- |
| `match.updated` | `matches` | Qualquer mudança de placar/status |
| `match.finished` | `matches` | Status virou `finished` |
| `match.goal` | `match:<matchID>` | Novo gol detectado |
| `ranking.updated` | `rankings` + `group:<groupID>` | Partida finalizada e pontos calculados |

---

## 4. Sincronização de partidas (MatchSync)

O worker `cmd/matchsync` faz polling na football-data.org e propaga mudanças via WebSocket.

```mermaid
flowchart TD
    START["Worker inicia"] --> SCHED["Scheduler detecta\ntipo de sync"]

    SCHED -- "ao vivo (a cada 30s)" --> LIVE["Fetch /matches?status=LIVE"]
    SCHED -- "jogos do dia (a cada 5min)" --> TODAY["Fetch /matches?dateFrom=hoje"]
    SCHED -- "próximos (a cada 1h)" --> UPCOMING["Fetch /matches?dateTo=+30d"]

    LIVE --> RATE["Rate limiter\n(min 6s entre chamadas)"]
    TODAY --> RATE
    UPCOMING --> RATE

    RATE --> FETCH["GET api.football-data.org/v4\n/competitions/WC/matches\nX-Auth-Token: <token>"]
    FETCH --> PARSE["Decode JSON → []ProviderMatch"]

    PARSE --> LOOP["Para cada partida"]

    LOOP --> UPSERT["UpsertProviderMatch\n(INSERT ON CONFLICT)"]
    UPSERT --> CHANGED{"changedRows > 0?"}

    CHANGED -- Não --> NEXT["próxima partida"]
    CHANGED -- Sim --> EVT_MATCH["Publica match.updated"]

    EVT_MATCH --> FINISHED{"status =\nfinished?"}
    FINISHED -- Sim --> EVT_FIN["Publica match.finished"]
    FINISHED -- Não --> GOALS["Sync gols"]

    EVT_FIN --> SCORE["ScoreProviderMatchPredictions\n(recalcula pontos)"]
    SCORE --> SCORED{"predictions\npontuadas?"}
    SCORED -- Sim --> EVT_RANK["Publica ranking.updated\npor grupo"]
    SCORED -- Não --> GOALS

    GOALS --> NEW_GOALS{"Novos gols?"}
    NEW_GOALS -- Sim --> EVT_GOAL["Publica match.goal\n(por gol, room match:<id>)"]
    NEW_GOALS -- Não --> NEXT

    EVT_GOAL --> NEXT
    EVT_RANK --> NEXT
    NEXT --> LOOP
```

---

## 5. Leitura de previsão de partida

O endpoint `GET /api/v1/matches/{matchID}/prediction` agrega probabilidades, expected goals, top placares e explicação da PalpitAI em uma única resposta.

```mermaid
sequenceDiagram
    participant App as App Mobile
    participant Controller as GetMatchPredictionHandler
    participant Reader as PredictionReadService
    participant DB as PostgreSQL

    App->>Controller: GET /api/v1/matches/{matchID}/prediction
    Note right of App: Authorization: Bearer <token>

    Controller->>Controller: valida token via Supabase

    Controller->>Reader: MatchPredictionByMatchID(matchID)
    Reader->>DB: SELECT FROM match_predictions
    DB-->>Reader: MatchPrediction (probabilidades)

    Controller->>Reader: GoalPredictionByMatchID(matchID)
    Reader->>DB: SELECT FROM match_goal_predictions + match_score_probabilities
    DB-->>Reader: MatchGoalPrediction (xG, placar mais provável, top scores)

    Controller->>Reader: ExplanationByMatchID(matchID, promptVersion)
    Reader->>DB: SELECT FROM prediction_explanations WHERE status = 'generated'
    DB-->>Reader: PredictionExplanation (opcional)

    Controller-->>App: 200 OK { match_id, probabilities, goals?, top_scores?, explanation? }
```

**Comportamento por campo:**

| Campo | Obrigatório | Comportamento se ausente |
| --- | --- | --- |
| `probabilities` | Sim | 404 se `match_prediction` não existir |
| `goals` | Não | Campo omitido na resposta |
| `top_scores` | Não | Campo omitido na resposta |
| `explanation` | Não | Campo omitido se não gerado ainda |

---

## 6. Geração de explicações da PalpitAI

O worker `cmd/workers/generate_prediction_explanations` lê previsões do ML e gera explicações em português via Gemini API.

```mermaid
sequenceDiagram
    participant Worker
    participant DB as PostgreSQL
    participant PromptBuilder
    participant Gemini

    Worker->>DB: FindPendingMatchesForExplanation(fromDate, toDate, limit, promptVersion)
    DB-->>Worker: []ExplanationCandidate

    loop Para cada candidato
        Worker->>PromptBuilder: BuildPromptInput(candidate, promptVersion)
        PromptBuilder->>PromptBuilder: valida campos obrigatórios
        alt Campos faltando
            Worker->>DB: MarkSkipped(matchID, motivo)
        else Candidato válido
            PromptBuilder-->>Worker: ExplanationPromptInput

            Worker->>Gemini: POST /v1beta/models/{model}:generateContent\n{ responseMimeType, responseSchema, prompt }
            Gemini-->>Worker: candidates.content.parts[0].text (JSON estruturado)

            Worker->>Worker: ParseAndValidateExplanation(output_text)
            alt JSON inválido
                Worker->>Gemini: retry com correction prompt (até 2x)
            end

            Worker->>DB: UpsertExplanation(matchID, summary, mainReasons,\nbetStyle, riskAlert, userTip, rawResponse)
            Note right of DB: status = 'generated'
        end
    end

    Worker->>Worker: Imprime resumo\n(processadas, geradas, puladas, falhas)
```

### Estrutura do prompt e output

**Input para o modelo (por partida):**
```
Match: Brasil vs Argentina, 2026-06-15
Result: HOME_WIN 62% | DRAW 21% | AWAY_WIN 17% (confidence: high)
Goals: xG casa 1.8 | xG fora 1.1 | placar mais provável 1x0
Top scores: 1x0 (18%), 2x0 (12%), 2x1 (10%)
Metrics: elo_diff=+85, form_home=72, form_away=64, wc_history_home=88
```

**Output estruturado (JSON schema validado):**
```json
{
  "summary": "Brasil entra como favorito com histórico superior...",
  "main_reasons": [
    "Vantagem de ELO significativa (+85)",
    "Melhor forma recente (72 vs 64)",
    "Histórico de Copa dominante"
  ],
  "bet_style": "moderate",
  "risk_alert": "Argentina tem atacantes de alto nível capazes de virar",
  "user_tip": "Aposte no Brasil vencendo, mas considere margem estreita"
}
```
