# Palpite! Frontend

App mobile em React Native com Expo para criação de grupos de bolão, palpites da Copa do Mundo 2026, ranking em tempo real e visualização de previsões de ML.

## O que é

O frontend é o ponto de entrada do usuário no Palpite!. Permite criar e entrar em grupos, registrar palpites antes do início de cada partida e acompanhar o ranking atualizado em tempo real via WebSocket. O app também exibe as previsões de resultado e placar geradas pelo pipeline de ML com explicações em linguagem natural.

## Tecnologias

- **React Native 0.81 + Expo 54** — framework mobile multiplataforma
- **TypeScript** — tipagem estática
- **Supabase Auth** (`@supabase/supabase-js`) — autenticação e sessão
- **React Query** (`@tanstack/react-query`) — cache e revalidação de dados
- **WebSocket** — atualizações em tempo real de jogos e ranking
- **ESLint + Prettier** — qualidade e formatação
- **Vitest** — testes unitários
- **Husky + Commitlint** — hooks de commit

## Fontes de dados

| Fonte            | Uso                                              |
| ---------------- | ------------------------------------------------ |
| Backend Go (API) | Grupos, palpites, ranking e previsões de ML      |
| Supabase Auth    | Login, cadastro, sessão e token de acesso        |
| WebSocket (/ws)  | Eventos em tempo real de jogos e ranking         |

## Configuração

```bash
npm install
cp .env.example .env
```

```env
EXPO_PUBLIC_SUPABASE_URL=https://project.supabase.co
EXPO_PUBLIC_SUPABASE_KEY=chave_publica
EXPO_PUBLIC_API_URL=http://SEU_IP_LOCAL:3000
```

Em dispositivo físico, `EXPO_PUBLIC_API_URL` deve usar o IP da máquina na rede local. `localhost` aponta para o próprio celular e não acessa a API.

## Como rodar

```bash
npm run start        # inicia o Expo (escolha a plataforma no terminal)
npm run android      # Android Emulator ou dispositivo
npm run ios          # iOS Simulator ou dispositivo
npm run web          # navegador
```

## Estrutura

```text
src/
├── features/
│   ├── auth/         # login, cadastro, sessão e Supabase Auth
│   ├── groups/       # Home, criação, detalhe, admin, palpites e ranking
│   ├── onboarding/   # fluxo inicial do app
│   └── realtime/     # conexão WebSocket e notificações
├── navigation/       # fluxo de telas (AppNavigator)
├── services/
│   └── supabase.ts   # cliente Supabase global
└── shared/
    ├── components/   # componentes reutilizáveis de UI
    ├── hooks/        # hooks globais
    ├── query/        # setup do React Query
    ├── services/     # API client HTTP
    └── theme/        # estilos e tema
```

## Fluxo principal

1. `App.tsx` monta os providers globais (QueryClient, Auth, Navigation)
2. `AuthProvider` controla a sessão do Supabase
3. `AppNavigator` decide a tela ativa com base no estado de onboarding e sessão
4. Telas das features usam hooks e services para buscar e mutar dados
5. Eventos WebSocket invalidam queries e exibem notificações sem refresh manual

## Exclusão de conta

O usuário autenticado pode solicitar a exclusão em Perfil > Configurações > Excluir conta. O app exige a confirmação digitada `EXCLUIR`, chama `DELETE /api/v1/me`, encerra a sessão e volta para o fluxo de login.

## Qualidade

```bash
npm run lint
npm run format:check
npm run typecheck
npm run test
```

## Commits

O projeto usa Conventional Commits com Husky e Commitlint:

```bash
feat: add signup flow
fix: handle private group approval
refactor(frontend): move navigation to app navigator
```
