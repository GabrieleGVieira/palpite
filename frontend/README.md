# PalpitAI Frontend

App mobile em React Native com Expo para criacao de grupos de bolao, palpites da Copa do Mundo, ranking e atualizacoes em tempo real.

## Stack

- React Native
- Expo
- TypeScript
- Supabase Auth
- React Query
- WebSocket
- ESLint
- Prettier
- Vitest
- Husky
- Commitlint

## Configuracao

```bash
npm install
cp .env.example .env
```

Variaveis:

```bash
EXPO_PUBLIC_SUPABASE_URL=https://project.supabase.co
EXPO_PUBLIC_SUPABASE_KEY=cole_a_chave_publica_aqui
EXPO_PUBLIC_API_URL=http://SEU_IP_LOCAL:3000
```

Ao rodar em dispositivo fisico, use o IP local da maquina no `EXPO_PUBLIC_API_URL`. `localhost` aponta para o proprio celular/emulador e normalmente nao acessa a API da maquina.

## Como rodar

```bash
npm run start
```

Plataformas:

```bash
npm run android
npm run ios
npm run web
```

## Estrutura

```text
src/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── screens/
│   │   ├── services/
│   │   └── store/
│   ├── groups/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── screens/
│   │   ├── services/
│   │   ├── utils/
│   │   └── types.ts
│   ├── onboarding/
│   │   └── screens/
│   └── realtime/
├── navigation/
│   ├── AppNavigator.tsx
│   └── types.ts
├── services/
│   └── supabase.ts
└── shared/
    ├── components/
    ├── hooks/
    ├── query/
    ├── services/
    └── theme/
```

## Organizacao

- `features/auth`: login, cadastro, sessao e Supabase Auth.
- `features/groups`: Home, criacao de grupo, detalhe do grupo, admin, palpites e ranking.
- `features/onboarding`: fluxo inicial do app.
- `features/realtime`: conexao WebSocket, tipos e notificacoes.
- `navigation`: regras de fluxo entre onboarding, auth, home, grupo e admin.
- `shared`: componentes, hooks, tema, API client e React Query compartilhados.
- `services/supabase.ts`: cliente Supabase global.

## Fluxo principal

1. `App.tsx` monta providers globais.
2. `AuthProvider` controla sessao do Supabase.
3. `QueryClientProvider` centraliza cache/revalidacao.
4. `AppNavigator` decide a tela ativa conforme onboarding, sessao e grupo selecionado.
5. Telas das features usam hooks locais e services para buscar dados.
6. Eventos realtime invalidam queries e exibem notificacoes sem refresh manual.

## Scripts

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
npm run test
```

## Qualidade

Antes de abrir PR ou commitar alteracoes relevantes:

```bash
npm run format:check
npm run lint
npm run typecheck
npm run test
```

## Commits

O projeto usa Husky e Commitlint. Use Conventional Commits:

```bash
feat: add signup flow
fix: handle private group approval
refactor(frontend): move navigation to app navigator
```
