# Palpite!

App mobile de bolão da Copa do Mundo 2026 com grupos, palpites, ranking em tempo real e previsões geradas por machine learning.

## O que é

Palpite! é um bolão social onde usuários criam ou entram em grupos, registram palpites para cada partida da Copa do Mundo e acompanham o ranking em tempo real. O app combina uma API em Go com um pipeline de ML em Python para gerar previsões de resultado e placar com explicações em linguagem natural via LLM.

## Arquitetura

```text
palpitAI/
├── backend/      # API HTTP/WebSocket em Go
├── frontend/     # App mobile React Native + Expo
├── ml-service/   # Pipeline ML e API de inferência em Python
└── docs/         # Documentação técnica dos motores de IA e métricas
```

## Stack

| Camada       | Tecnologias                                              |
| ------------ | -------------------------------------------------------- |
| Mobile       | React Native, Expo, TypeScript, Supabase Auth, React Query |
| API          | Go, net/http, WebSocket, Redis                           |
| ML           | Python, scikit-learn, FastAPI, pandas                    |
| Banco        | PostgreSQL (Supabase)                                    |
| Cache        | Redis (Upstash)                                          |
| IA           | Gemini API para explicações                              |
| Dados        | football-data.org, CSVs históricos de partidas internacionais |

## Requisitos

- Go 1.24+
- Node.js + npm
- Python 3.11+
- Projeto Supabase (PostgreSQL)
- Redis (Upstash)
- Token da [football-data.org](https://www.football-data.org/)
- Chave da Gemini API
- Expo Go, Android Emulator ou iOS Simulator

## Configuração

**Backend:**

```bash
cd backend
cp .env.example .env
```

```env
APP_ENV=development
PORT=3000
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres
REDIS_URL=redis://default:token@host.upstash.io:6379
SUPABASE_URL=https://project.supabase.co
SUPABASE_KEY=chave_publica
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

**Frontend:**

```bash
cd frontend
cp .env.example .env
```

```env
EXPO_PUBLIC_SUPABASE_URL=https://project.supabase.co
EXPO_PUBLIC_SUPABASE_KEY=chave_publica
EXPO_PUBLIC_API_URL=http://SEU_IP_LOCAL:3000
```

Em dispositivo físico, `EXPO_PUBLIC_API_URL` deve usar o IP da máquina na rede local.

**ML Service:**

```bash
cd ml-service
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
export DATABASE_URL='postgresql://...'
```

## Como rodar

**API:**

```bash
cd backend
make run
```

**Worker de sincronização de partidas:**

```bash
cd backend
go run ./cmd/matchsync
```

**App mobile:**

```bash
cd frontend
npm install
npm run start
```

**ML Service (API de inferência):**

```bash
cd ml-service
uvicorn app.api.main:app --reload
```

## Funcionalidades

- Autenticação via Supabase Auth (login, cadastro, logout, sessão)
- Criação e entrada em grupos de bolão com código de convite
- Grupos públicos e privados com aprovação pelo owner
- Palpites por partida com pontuação automática
- Pontuação: 10 pts para placar exato, 7 pts para vencedor/empate e o número de gols de um dos times correto, 5 pts Mesmo vencedor ou empate (sem acertar nenhum gol exato), 3 pts Errou o vencedor, mas acertou a quantidade de gols de um dos times e 0 pts para erro
- Ranking em tempo real por grupo via WebSocket
- Sincronização automática de placares via football-data.org
- Previsões de resultado e placar geradas por ML
- Explicações das previsões em linguagem natural via LLM

## Qualidade

**Backend:**

```bash
cd backend
make fmt && make vet && make test
```

**Frontend:**

```bash
cd frontend
npm run lint && npm run typecheck && npm run test
```

## Commits

O projeto usa Conventional Commits com Husky e Commitlint:

```bash
feat: add group ranking
fix: correct prediction score
refactor(frontend): organize app by feature folders
```
