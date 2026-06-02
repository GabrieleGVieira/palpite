# Palpite!

App mobile e PWA de bolão da Copa do Mundo 2026 com grupos, palpites, ranking em tempo real, feed social, amigos, desafios com Palpicoins e análises complementares da PalpitAI.

## O que é

Palpite! é uma plataforma social de bolões e previsões esportivas onde Palpiteiros criam ou entram em grupos, registram palpites para cada partida da Copa do Mundo e acompanham rankings em tempo real. A experiência também inclui perfil com avatar, amigos, perfis públicos, feed de atividades por grupo, controle de pagamentos em bolões pagos, desafios entre amigos e carteira de Palpicoins. A PalpitAI aparece como recurso complementar, com análises, sugestões e explicações em linguagem natural.

## Arquitetura

```text
palpite/
├── backend/      # API HTTP/WebSocket em Go
├── frontend/     # App mobile React Native + Expo
├── landing/      # Landing page, páginas legais e PWA web
├── ml-service/   # Pipeline ML e API de inferência em Python
└── docs/         # Documentação técnica dos motores da PalpitAI e métricas
```

## Stack

| Camada       | Tecnologias                                              |
| ------------ | -------------------------------------------------------- |
| Mobile       | React Native, Expo, TypeScript, Supabase Auth, React Query |
| Web pública  | React, Vite, TypeScript, Supabase Auth                    |
| API          | Go, net/http, WebSocket, Redis                           |
| ML           | Python, scikit-learn, FastAPI, pandas                    |
| Banco        | PostgreSQL (Supabase)                                    |
| Cache        | Redis (Upstash)                                          |
| PalpitAI     | Gemini API para explicações e análises complementares     |
| Dados        | football-data.org, CSVs históricos de partidas internacionais |

## Requisitos

- Go 1.24+
- Node.js 22.12+ + npm
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
AI_EXPLANATION_BATCH_SIZE=2
AI_EXPLANATION_MIN_BATCH_SIZE=1
AI_EXPLANATION_RETRY_MISSING=true
AI_EXPLANATION_MAX_MISSING_RETRIES=2
AI_EXPLANATION_SEED_DAYS=90
AI_EXPLANATION_REFRESH_DAYS=7
AI_EXPLANATION_MAX_AGE_HOURS=24
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

**Landing/PWA:**

```bash
cd landing
cp .env.example .env
npm install
```

```env
VITE_API_URL=http://SEU_IP_LOCAL:3000
VITE_SUPABASE_URL=https://project.supabase.co
VITE_SUPABASE_ANON_KEY=chave_publica
```

`VITE_SUPABASE_KEY` ainda é aceito por compatibilidade, mas `VITE_SUPABASE_ANON_KEY` é o nome preferido.

## Como rodar

**API:**

```bash
cd backend
make run
```

**Worker de sincronização de partidas:**

```bash
cd backend
make matchsync
```

**App mobile:**

```bash
cd frontend
npm install
npm run start
```

**Landing/PWA:**

```bash
cd landing
npm install
npm run dev
```

**ML Service (API de inferência):**

```bash
cd ml-service
uvicorn app.api.main:app --reload
```

## Funcionalidades

- Autenticação via Supabase Auth (login, cadastro, logout, sessão)
- Perfil editável com nome, avatar em Supabase Storage e visibilidade pública
- Criação e entrada em grupos de bolão com código de convite
- Grupos públicos e privados com aprovação pelo owner
- Administração de grupo com edição, aprovação de entrada, remoção de membros, transferência de ownership e saída do grupo
- Bolões pagos com valor configurável, resumo de pagamentos e marcação manual de status por participante
- Palpites por partida com pontuação automática
- Pontuação: 10 pts para placar exato, 7 pts para vencedor/empate e o número de gols de um dos times correto, 5 pts Mesmo vencedor ou empate (sem acertar nenhum gol exato), 3 pts Errou o vencedor, mas acertou a quantidade de gols de um dos times e 0 pts para erro
- Ranking em tempo real por grupo via WebSocket
- Feed de atividades do grupo com reações
- Busca de usuários, solicitações de amizade, lista de amigos e perfil público
- Desafios entre amigos com stake em Palpicoins
- Carteira, histórico e ranking global de Palpicoins
- Sincronização automática de placares via football-data.org
- Previsões de resultado e placar geradas por ML
- Análises da PalpitAI em linguagem natural via LLM
- Landing pública com privacidade, termos, exclusão de conta e formulário Beta Android
- PWA gerada pelo app Expo em `frontend` com manifest e service worker preparados no build web

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

**Landing:**

```bash
cd landing
npm run build
```

**ML Service:**

```bash
cd ml-service
pytest
```

## Commits

O projeto usa Conventional Commits com Husky e Commitlint:

```bash
feat: add group ranking
fix: correct prediction score
refactor(frontend): organize app by feature folders
```
