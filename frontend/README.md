# Palpite! Frontend

App mobile em React Native com Expo para criação de grupos de bolão, palpites da Copa do Mundo 2026, ranking em tempo real, feed social, amigos, desafios com Palpicoins e visualização de análises da PalpitAI.

## O que é

O frontend é o ponto de entrada dos Palpiteiros no Palpite! O app permite criar e entrar em grupos, registrar palpites antes do início de cada partida, acompanhar ranking atualizado em tempo real via WebSocket, editar perfil com avatar, interagir com feed de grupo, gerenciar pagamentos, fazer amigos, consultar perfis públicos e criar desafios usando Palpicoins. O app também exibe análises da PalpitAI geradas pelo pipeline de ML com explicações em linguagem natural.

## Tecnologias

- **React Native 0.81 + Expo 54** — framework mobile multiplataforma
- **TypeScript** — tipagem estática
- **Supabase Auth** (`@supabase/supabase-js`) — autenticação e sessão
- **React Query** (`@tanstack/react-query`) — cache e revalidação de dados
- **WebSocket** — atualizações em tempo real de jogos e ranking
- **Expo Image Picker + FileSystem** — seleção e upload de avatar
- **Supabase Storage** — armazenamento público dos avatares
- **ESLint + Prettier** — qualidade e formatação
- **Vitest** — testes unitários
- **Husky + Commitlint** — hooks de commit

## Fontes de dados

| Fonte            | Uso                                              |
| ---------------- | ------------------------------------------------ |
| Backend Go (API) | Grupos, palpites, ranking e análises de ML      |
| Supabase Auth    | Login, cadastro, sessão e token de acesso        |
| WebSocket (/ws)  | Eventos em tempo real de jogos e ranking         |

## Configuração

```bash
cd frontend
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

## PWA web

O app Expo também pode ser exportado como PWA:

```bash
npm run build:web
```

O script gera `dist/` com o export web do Expo e roda `scripts/prepare-expo-web-pwa.mjs` para copiar manifest, service worker e assets PWA. Veja `PWA_SETUP.md` para notas de publicação e instalação.

## Estrutura

```text
src/
├── features/
│   ├── auth/         # login, cadastro, sessão e Supabase Auth
│   ├── account/      # perfil, avatar, privacidade e exclusão de conta
│   ├── challenges/   # desafios entre amigos com Palpicoins
│   ├── friends/      # busca, amizades e perfis públicos
│   ├── groups/       # Home, criação, detalhe, admin, membros, feed, pagamentos e ranking
│   ├── onboarding/   # fluxo inicial do app
│   ├── palpicoins/   # carteira, histórico e ranking global
│   ├── predictions/  # cards da PalpitAI e sugestão de placar
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

## Funcionalidades

- Login, cadastro, logout e sessão persistente via Supabase Auth
- Perfil com nome público, avatar em Supabase Storage e opção de perfil público
- Home com grupos, pontuação total, atalhos para amigos, desafios, Palpicoins e perfil
- Criação de grupo com privacidade, limite de participantes, escopo de partidas, bolão pago e bloqueio de palpites pendentes
- Entrada por código de convite, com aprovação do owner em grupos privados
- Administração de grupo com edição, aprovação de entrada, lista de membros, remoção, saída, transferência de ownership e controle de pagamentos
- Detalhe do grupo com tabs de jogos, ranking e feed de atividades com reações
- Palpites por partida antes do kickoff, sugestão de placar a partir da PalpitAI e atualização de cache após salvar
- Busca de usuários, solicitações de amizade, lista de amigos e perfil público com estatísticas
- Desafios entre amigos com stake em Palpicoins e estados `PENDING`, `ACCEPTED`, `DECLINED`, `CANCELLED` e `SETTLED`
- Carteira de Palpicoins, histórico de transações e ranking global
- Exclusão de conta com confirmação textual

## Perfil, avatar e exclusão de conta

O usuário autenticado pode atualizar nome, avatar e visibilidade pública em Perfil. O upload usa o bucket público `avatars` do Supabase Storage, com políticas em `docs/supabase-avatar-storage.sql`.

A exclusão fica em Perfil > Configurações > Excluir conta. O app exige a confirmação digitada `EXCLUIR`, chama `DELETE /api/v1/me`, encerra a sessão e volta para o fluxo de login.

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
