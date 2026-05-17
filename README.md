# PalpitAI

O bolao inteligente: app mobile de Copa do Mundo com grupos privados, rankings, palpites em tempo real e insights de IA.

## Stack

| Camada | Tecnologias |
| ------ | ----------- |
| Mobile | React Native, Expo, TypeScript, Supabase Auth |
| API | Go, net/http |
| Qualidade | ESLint, Prettier, Husky, Commitlint |

## Estrutura

```text
palpitAI/
├── backend/           # API em Go
│   ├── cmd/api/
│   ├── internal/
│   ├── go.mod
│   └── Makefile
├── frontend/          # App Expo React Native
│   ├── App.tsx
│   ├── index.ts
│   ├── app.json
│   ├── eslint.config.js
│   ├── commitlint.config.js
│   └── package.json
├── codealike.json
└── README.md
```

## Requisitos

- Node.js
- npm
- Go 1.24+
- Expo Go no dispositivo fisico, Android Emulator ou iOS Simulator

## Como rodar

### Backend

```bash
cd backend
cp .env.example .env
make run
```

A API inicia em `http://localhost:3000`.

Rotas iniciais:

```text
GET /health
GET /api/v1/status
GET /api/v1/groups
POST /api/v1/groups
PUT /api/v1/groups/{groupID}
POST /api/v1/groups/join
GET /api/v1/groups/{groupID}/matches
PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction
```

### Frontend

Instale as dependencias:

```bash
cd frontend
npm install
```

Configure o Supabase Auth:

```bash
cp .env.example .env
```

Preencha `EXPO_PUBLIC_SUPABASE_KEY` com a chave publica do projeto Supabase e `EXPO_PUBLIC_API_URL` com a URL do backend.

Inicie o Metro Bundler:

```bash
npm run start
```

Ou rode diretamente em uma plataforma:

```bash
npm run android
npm run ios
npm run web
```

## Qualidade de codigo

```bash
npm run lint          # Executa ESLint
npm run lint:fix      # Corrige problemas automaticos do ESLint
npm run format        # Formata com Prettier
npm run format:check  # Verifica formatacao
npm run typecheck     # Verifica TypeScript
```

## Commits

O projeto usa Husky e Commitlint para validar commits no padrao Conventional Commits.

Exemplos:

```bash
feat: add login screen
fix: correct prediction score
chore: update dependencies
```

Hooks configurados:

- `pre-commit`: roda `npm run lint` e `npm run typecheck`
- `commit-msg`: valida a mensagem com Commitlint

## Scripts do frontend

```bash
npm run start
npm run android
npm run ios
npm run web
npm run lint
npm run lint:fix
npm run format
npm run format:check
npm run typecheck
```

## Scripts do backend

```bash
make run
make test
make fmt
make vet
```
