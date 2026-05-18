# PalpitAI

App mobile de bolao da Copa do Mundo com grupos, palpites, ranking em tempo real e base preparada para recursos de IA.

## Stack

| Camada | Tecnologias |
| --- | --- |
| Mobile | React Native, Expo, TypeScript, Supabase Auth, React Query |
| API | Go, net/http, WebSocket |
| Banco | Supabase/Postgres |
| Realtime | WebSocket no backend + sync de placares via football-data.org |
| Qualidade | ESLint, Prettier, Vitest, Husky, Commitlint |

## Estrutura

```text
palpitAI/
├── backend/
│   ├── cmd/
│   │   ├── api/          # API HTTP/WebSocket
│   │   ├── matchsync/    # worker separado de sincronizacao de jogos
│   │   └── seed/         # carga inicial de dados
│   ├── docs/
│   ├── internal/
│   │   ├── controller/
│   │   ├── domain/
│   │   ├── dto/
│   │   ├── matchsync/
│   │   ├── realtime/
│   │   ├── repositories/
│   │   ├── route/
│   │   └── usecase/
│   └── Makefile
├── frontend/
│   ├── App.tsx
│   ├── src/
│   │   ├── features/     # auth, groups, onboarding, realtime
│   │   ├── navigation/   # fluxo de telas do app
│   │   ├── services/     # clientes globais, como Supabase
│   │   └── shared/       # UI, hooks, query, theme e api client
│   └── package.json
└── README.md
```

## Requisitos

- Node.js
- npm
- Go 1.24+
- Conta/projeto Supabase
- Token do football-data.org para sincronizacao automatica dos jogos
- Expo Go, Android Emulator ou iOS Simulator para rodar o app

## Configuracao

Backend:

```bash
cd backend
cp .env.example .env
```

Variaveis principais:

```bash
APP_ENV=development
PORT=3000
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres
SUPABASE_URL=https://project.supabase.co
SUPABASE_KEY=cole_a_chave_publica_aqui
FOOTBALL_DATA_API_BASE_URL=https://api.football-data.org/v4
FOOTBALL_DATA_COMPETITION_CODE=WC
FOOTBALL_DATA_SEASON=2026
FOOTBALL_DATA_TOKEN=cole_o_token_aqui
```

Frontend:

```bash
cd frontend
cp .env.example .env
```

Variaveis principais:

```bash
EXPO_PUBLIC_SUPABASE_URL=https://project.supabase.co
EXPO_PUBLIC_SUPABASE_KEY=cole_a_chave_publica_aqui
EXPO_PUBLIC_API_URL=http://SEU_IP_LOCAL:3000
```

Em dispositivo fisico, `EXPO_PUBLIC_API_URL` deve usar o IP da maquina na rede local, nao `localhost`.

## Como rodar

API:

```bash
cd backend
make run
```

Worker de sincronizacao de partidas:

```bash
cd backend
go run ./cmd/matchsync
```

App:

```bash
cd frontend
npm install
npm run start
```

## Funcionalidades atuais

- Onboarding, login, cadastro, logout e sessao via Supabase Auth.
- Criacao de grupos de bolao da Copa do Mundo.
- Entrada em grupos via codigo de convite.
- Grupos publicos e privados, com aceite pelo owner quando privado.
- Tela Home com grupos do usuario, pontuacao geral e pendencias de aprovacao.
- Tela do grupo com jogos, palpites e ranking.
- Tela admin do grupo para editar informacoes e aceitar solicitacoes.
- Pontuacao fixa: 10 pontos para placar exato, 5 para vencedor/empate, 0 para erro.
- Sincronizacao de placares via football-data.org.
- WebSocket para atualizar jogos, ranking e notificacoes sem refresh manual.

## Principais rotas da API

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

## Qualidade

Frontend:

```bash
cd frontend
npm run lint
npm run format:check
npm run typecheck
npm run test
```

Backend:

```bash
cd backend
make fmt
make vet
make test
```

## Commits

O projeto usa Husky e Commitlint com Conventional Commits.

Exemplos:

```bash
feat: add group ranking
fix: correct prediction score
refactor(frontend): organize app by feature folders
```
