# Frontend — Fluxo do Sistema

Documentação dos fluxos principais do app React Native/Expo do Palpite!: inicialização, autenticação, grupos, palpites, feed, amigos, Palpicoins, desafios, perfil, realtime e export PWA.

---

## Visão geral

```mermaid
graph TD
    APP["App.tsx\n(providers globais)"]
    AUTH["AuthProvider\n(sessão Supabase)"]
    QCP["QueryClientProvider\n(React Query cache)"]
    NAV["AppNavigator\n(roteamento)"]

    APP --> QCP
    APP --> AUTH
    APP --> NAV

    NAV --> ONBOARD["OnboardingScreen\n(primeira vez)"]
    NAV --> LOGIN["LoginScreen"]
    NAV --> SIGNUP["SignupScreen"]
    NAV --> HOME["HomeScreen"]
    NAV --> CREATE["CreateGroupScreen"]
    NAV --> DETAIL["GroupDetailScreen"]
    NAV --> ADMIN["GroupAdminScreen"]
    NAV --> FRIENDS["FriendsScreen"]
    NAV --> PROFILE["ProfileScreen"]
    NAV --> COINS["PalpicoinsScreen"]
    NAV --> CHALLENGES["ChallengesScreen"]

    HOME -- "WebSocket" --> REALTIME["useRealtimeEvents"]
    DETAIL -- "WebSocket" --> REALTIME
    REALTIME --> CACHE["Invalida cache\nReact Query"]

    DETAIL -- "abre partida" --> PRED["MatchPredictionCard\n(useMatchPrediction)"]
    PRED -- "GET /api/v1/matches/{id}/prediction" --> BACKEND["Backend Go"]
```

---

## 1. Inicialização do app

```mermaid
sequenceDiagram
    participant App as App.tsx
    participant Auth as AuthProvider
    participant Supa as Supabase
    participant Nav as AppNavigator

    App->>Auth: monta AuthProvider
    Auth->>Supa: supabase.auth.getSession()
    Supa-->>Auth: Session | null

    Auth->>Auth: setSession(), isLoading = false
    Auth->>Supa: onAuthStateChange (subscribe)

    App->>Nav: renderiza AppNavigator

    alt isLoading = true
        Nav->>Nav: exibe ActivityIndicator
    else session = null
        Nav->>Nav: exibe fluxo de Auth
        Note right of Nav: OnboardingScreen (se primeira vez)\n→ LoginScreen | SignupScreen
    else session existe
        Nav->>Nav: exibe fluxo do App
        Note right of Nav: HomeScreen (padrão)\nProfile/Friends/Palpicoins/Challenges\nGroupDetail/Admin/Members
    end
```

---

## 2. Fluxo de autenticação

```mermaid
flowchart TD
    ENTRY["App abre sem sessão"] --> OB{"Já viu\nonboarding?"}
    OB -- Não --> ONBOARD["OnboardingScreen"]
    ONBOARD --> LOGIN_SCREEN["LoginScreen"]
    OB -- Sim --> LOGIN_SCREEN

    LOGIN_SCREEN --> FILL_LOGIN["Preenche email + senha"]
    FILL_LOGIN --> DO_LOGIN["supabase.auth.signInWithPassword()"]
    DO_LOGIN --> LOGIN_OK{"Sucesso?"}
    LOGIN_OK -- Não --> LOGIN_ERR["Exibe erro"]
    LOGIN_ERR --> LOGIN_SCREEN
    LOGIN_OK -- Sim --> SESSION["Sessão criada\n→ HomeScreen"]

    LOGIN_SCREEN --> GOTO_SIGNUP["Ir para Signup"]
    GOTO_SIGNUP --> SIGNUP_SCREEN["SignupScreen"]
    SIGNUP_SCREEN --> FILL_SIGNUP["Preenche nome, email,\nsenha, confirmação"]
    FILL_SIGNUP --> DO_SIGNUP["supabase.auth.signUp()\n{ full_name: name }"]
    DO_SIGNUP --> SIGNUP_OK{"Sucesso?"}
    SIGNUP_OK -- Não --> SIGNUP_ERR["Exibe erro"]
    SIGNUP_ERR --> SIGNUP_SCREEN
    SIGNUP_OK -- Sim --> SESSION

    SESSION --> LOGOUT["logout()"]
    LOGOUT --> SUPA_OUT["supabase.auth.signOut()"]
    SUPA_OUT --> CLEAR["queryClient.clear()"]
    CLEAR --> LOGIN_SCREEN
```

**Token de sessão:** o `access_token` do Supabase é extraído a cada chamada HTTP pelo `apiClient` e enviado no header `Authorization: Bearer <token>`.

---

## 3. Fluxo de grupos, membros, feed e palpites

```mermaid
flowchart TD
    HOME["HomeScreen"] --> FETCH_GROUPS["GET /api/v1/groups\nuseHomeData()"]
    HOME --> FETCH_SCORE["GET /api/v1/me/score"]
    HOME --> SHORTCUTS["Atalhos: Perfil, Amigos,\nPalpicoins, Desafios"]

    HOME --> JOIN_FORM["Formulário: código de convite"]
    JOIN_FORM --> DO_JOIN["POST /api/v1/groups/join\n{ invite_code }"]
    DO_JOIN --> JOIN_RESULT{"status?"}
    JOIN_RESULT -- "ativo" --> INVAL_GROUPS["Invalida cache ['groups']"]
    JOIN_RESULT -- "pendente" --> MSG["Exibe 'aguardando aprovação'"]

    HOME --> GO_CREATE["→ CreateGroupScreen"]
    GO_CREATE --> FORM_CREATE["Preenche nome, descrição,\nprivacidade, times, limite,\nbolão pago e bloqueio de pendentes"]
    FORM_CREATE --> DO_CREATE["POST /api/v1/groups"]
    DO_CREATE --> INVAL_GROUPS

    HOME --> SELECT_GROUP["Seleciona grupo\n→ GroupDetailScreen"]

    SELECT_GROUP --> TABS{"Tab ativa"}
    TABS -- "Jogos" --> MATCHES["GET /api/v1/groups/{gid}/matches"]
    TABS -- "Ranking" --> RANKING["GET /api/v1/groups/{gid}/ranking\n(lazy: só carrega ao clicar)"]
    TABS -- "Feed" --> FEED["GET /api/v1/groups/{gid}/feed"]

    MATCHES --> CARD["Match card\n(placar atual + palpite do Palpiteiro)"]
    CARD --> EDIT_PRED["Edita home_score / away_score\n(somente antes do kickoff)"]
    EDIT_PRED --> SAVE_PRED["PUT /api/v1/groups/{gid}/matches/{mid}/prediction"]
    SAVE_PRED --> UPDATE_CACHE["Atualiza cache in-place\nInvalida ranking + score"]

    CARD --> AI_PRED["MatchPredictionCard\n(visível só para partidas scheduled)"]
    AI_PRED --> FETCH_PRED["GET /api/v1/matches/{mid}/prediction\n(useMatchPrediction)"]
    FETCH_PRED --> AI_BOX["AiExplanationBox\nProbabilityBar\nExpectedGoalsRow\nTopScoresList"]
    AI_BOX --> SUGGEST["Sugerir placar\n(onUseSuggestion → preenche palpite)"]

    SELECT_GROUP --> IS_OWNER{"é owner?"}
    IS_OWNER -- Sim --> ADMIN_BTN["→ GroupAdminScreen"]
    ADMIN_BTN --> EDIT_GROUP["PUT /api/v1/groups/{gid}"]
    ADMIN_BTN --> REQUESTS["GET /api/v1/groups/{gid}/join-requests"]
    REQUESTS --> APPROVE["POST .../join-requests/{uid}/approve"]
    ADMIN_BTN --> PAYMENTS["GET /payments + /payments/summary\nPATCH /payments/{uid}"]
    SELECT_GROUP --> MEMBERS["GET /api/v1/groups/{gid}/members"]
    MEMBERS --> MEMBER_DETAIL["GET /api/v1/groups/{gid}/members/{uid}"]
    MEMBERS --> TRANSFER["POST /transfer-ownership"]
    MEMBERS --> REMOVE["DELETE /members/{uid}"]
    FEED --> REACTION["POST/DELETE /feed/{eventID}/reaction"]
```

---

## 4. Perfil, amigos, Palpicoins e desafios

```mermaid
flowchart TD
    HOME["HomeScreen"] --> PROFILE["ProfileScreen"]
    PROFILE --> GET_PROFILE["GET /api/v1/me/profile"]
    PROFILE --> UPDATE_PROFILE["PATCH /api/v1/me/profile\nnome, avatar_url, perfil público"]
    UPDATE_PROFILE --> AVATAR["uploadAvatar()\nSupabase Storage bucket avatars"]
    PROFILE --> DELETE_ACCOUNT["DELETE /api/v1/me\nconfirmação EXCLUIR"]

    HOME --> FRIENDS["FriendsScreen"]
    FRIENDS --> LIST_FRIENDS["GET /api/v1/friends"]
    FRIENDS --> REQUESTS["GET /api/v1/friends/requests"]
    FRIENDS --> SEARCH["GET /api/v1/users/search?q="]
    SEARCH --> SEND["POST /api/v1/friends/request"]
    REQUESTS --> ACCEPT["POST /friends/{id}/accept"]
    REQUESTS --> DECLINE["POST /friends/{id}/decline"]
    FRIENDS --> PUBLIC["PublicProfileScreen\nGET /users/{id}/profile"]

    HOME --> COINS["PalpicoinsScreen"]
    COINS --> WALLET["GET /api/v1/me/wallet"]
    COINS --> TX["GET /api/v1/me/wallet/transactions?limit=30"]
    COINS --> RANKING["GET /api/v1/rankings/palpicoins"]

    HOME --> CHALLENGES["ChallengesScreen"]
    CHALLENGES --> LIST_CH["GET /api/v1/challenges?status=&type="]
    PUBLIC --> CREATE_CH["ChallengeFriendModal\nPOST /api/v1/challenges"]
    LIST_CH --> ACTIONS["accept / decline / cancel"]
```

Palpicoins são exibidos no app como moeda virtual sem valor monetário. O frontend mostra o aviso retornado pela API nos fluxos de carteira, histórico, ranking e desafios.

---

## 4.1 PWA web

O mesmo app Expo pode ser executado no navegador com `npm run web` durante desenvolvimento ou exportado com:

```bash
cd frontend
npm run build:web
```

O build gera `frontend/dist` e executa `scripts/prepare-expo-web-pwa.mjs`, que prepara manifest, service worker e assets PWA para publicação.

---

## 5. Fluxo de requisições HTTP

Todas as chamadas ao backend passam pelo `apiClient`, que injeta o token automaticamente.

```mermaid
sequenceDiagram
    participant Hook as Hook/Service
    participant Client as apiClient
    participant Supa as Supabase
    participant API as Backend Go

    Hook->>Client: apiClient('/api/v1/groups', { method: 'GET' })
    Client->>Supa: supabase.auth.getSession()
    Supa-->>Client: { access_token }
    Client->>API: GET /api/v1/groups\nAuthorization: Bearer <token>\nContent-Type: application/json
    Note right of Client: AbortController timeout: 15s
    API-->>Client: 200 { groups: [...] }
    Client->>Client: parse JSON response
    Client-->>Hook: data
```

---

## 6. Fluxo realtime (WebSocket)

```mermaid
sequenceDiagram
    participant Hook as useRealtimeEvents
    participant WS as WebSocket
    participant Cache as React Query Cache
    participant UI

    Hook->>Hook: getSession() → access_token
    Hook->>WS: new WebSocket(wss://<api>/ws?token=<jwt>&group_id=<uuid>)
    WS-->>Hook: onopen (conectado)

    loop Evento do servidor
        WS-->>Hook: onmessage(event.data)
        Hook->>Hook: JSON.parse(event)

        alt match.updated
            Hook->>Cache: atualiza match in-place\n(home_score, away_score, status)
            Hook->>UI: showNotification("Placar atualizado")
        else match.finished
            Hook->>Cache: atualiza match (final_score, status, finished_at)
            Hook->>Cache: invalidate ['groups', gid, 'ranking']
            Hook->>Cache: invalidate ['me', 'score']
            Hook->>UI: showNotification("Brasil 2x1 Argentina - resultado final")
        else match.goal
            Hook->>Cache: atualiza match scores in-place
        else ranking.updated
            Hook->>Cache: invalidate ['groups', gid, 'ranking']
            Hook->>Cache: invalidate ['me', 'score']
            Hook->>UI: showNotification("Ranking atualizado")
        end
    end

    alt Conexão cai
        WS-->>Hook: onclose
        Hook->>Hook: setTimeout(reconnect, 2000ms)
    end

    Note over Hook: No unmount: isClosed=true, ws.close()
```

### Mapa de eventos × cache

| Evento | Payload relevante | Cache invalidado / atualizado |
| --- | --- | --- |
| `match.updated` | match_id, home_score, away_score, status | Atualiza in-place `['groups', gid, 'matches']` |
| `match.finished` | match_id, final scores, finished_at | Atualiza in-place + invalida ranking e score |
| `match.goal` | match_id, scores parciais | Atualiza in-place |
| `ranking.updated` | group_name | Invalida ranking e score |

---

## 7. Gerenciamento de estado

```mermaid
graph LR
    subgraph "React Query (server state)"
        G["['groups']"]
        S["['me', 'score']"]
        M["['groups', gid, 'matches']"]
        R["['groups', gid, 'ranking']"]
        J["['groups', gid, 'join-requests']"]
        P["['matchPrediction', matchID]"]
        F["['friends'] / ['friendRequests']"]
        W["['palpicoins', ...]"]
        C["['challenges', ...]"]
        FEED["['groups', gid, 'feed']"]
    end

    subgraph "Context (global)"
        AUTH["AuthProvider\n(session, user, login/logout)"]
    end

    subgraph "Local state (hooks)"
        DRAFTS["drafts[matchID]\n(palpites não salvos)"]
        TAB["tab ativa\n(matches | ranking)"]
        FORM["inputs de formulário\n(create/join/admin)"]
        NAV_STATE["appScreen + selectedGroup\n(navegação)"]
    end

    REALTIME["Eventos WebSocket"] -- "atualiza/invalida" --> G
    REALTIME -- "atualiza/invalida" --> S
    REALTIME -- "atualiza/invalida" --> M
    REALTIME -- "atualiza/invalida" --> R
```

**Configuração do React Query:**
- `staleTime: 15.000ms` — dados frescos por 15s, sem refetch desnecessário
- `refetchOnReconnect: true` — revalida ao reconectar
- Mutations com `retry: 0` — sem auto-retry em erros
- `['matchPrediction', matchID]` com `staleTime: 5min` — previsões mudam pouco; `retry: 1`

---

## 8. Feature de análises da PalpitAI

O módulo `features/predictions` exibe a previsão gerada pelo ML e PalpitAI diretamente no card de partida, antes do kickoff.

```mermaid
flowchart TD
    CARD["GroupDetailMatchCard\n(partida selecionada)"] --> STATUS{"status =\nscheduled?"}
    STATUS -- Não --> HIDE["MatchPredictionCard oculto"]
    STATUS -- Sim --> HOOK["useMatchPrediction(matchId, status)"]

    HOOK --> QUERY["GET /api/v1/matches/{matchID}/prediction\n(React Query cache 5min)"]

    QUERY --> LOADING["Estado: isLoading\n→ 'Carregando previsão...'"]
    QUERY --> ERROR["Estado: error\n→ 'Não foi possível carregar'"]
    QUERY --> NULL["Retorno: null (404)\n→ 'Previsão ainda não disponível'"]
    QUERY --> DATA["Retorno: MatchPrediction"]

    DATA --> PROB["ProbabilityBar\n(home_win / draw / away_win em %)"]
    DATA --> GOALS["ExpectedGoalsRow\n(xG casa, xG fora, placar mais provável)"]
    DATA --> TOP["TopScoresList\n(top placares com % individual)"]
    DATA --> EXPL{"explanation\npresente?"}

    EXPL -- Sim --> AI["AiExplanationBox\n(summary, main_reasons, risk_alert, user_tip)"]
    EXPL -- Não --> HIDE2["AiExplanationBox oculto"]

    DATA --> SUGGEST["Botão 'Usar sugestão'\nonUseSuggestion → preenche palpite"]
```

**Componentes do módulo:**

| Componente | Responsabilidade |
| --- | --- |
| `MatchPredictionCard` | Container principal; orquestra loading/error/empty states |
| `ProbabilityBar` | Barra visual de probabilidades home/draw/away |
| `ExpectedGoalsRow` | xG de cada time e placar mais provável |
| `TopScoresList` | Lista os top placares por probabilidade |
| `AiExplanationBox` | Exibe a explicação gerada pelo Gemini (condicional) |

**Regra de visibilidade:** `useMatchPrediction` só busca dados quando `isScheduledStatus(status) = true`. Para partidas live, finished ou timed, o card não é exibido.
