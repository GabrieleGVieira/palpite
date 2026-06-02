# Palpite! Landing

Landing page independente do app mobile, criada com React, TypeScript, Vite e CSS simples. Hospeda a página pública, páginas legais e o formulário Beta Android.

## Configuração

```bash
cd landing
cp .env.example .env
```

```env
VITE_API_URL=https://palpitai-api.onrender.com
VITE_SUPABASE_URL=https://project.supabase.co
VITE_SUPABASE_ANON_KEY=chave_publica
```

`VITE_SUPABASE_KEY` também é aceito por compatibilidade, mas `VITE_SUPABASE_ANON_KEY` é o nome preferido.

## Rodar localmente

```bash
cd landing
npm install
npm run dev
```

## Build

```bash
cd landing
npm run build
```

Para testar o build local:

```bash
npm run preview
```

## GitHub Pages

Por padrão o Vite usa `base: "/"`, adequado para domínio próprio como `https://palpite.app`.

Se o deploy for em um repositório GitHub Pages com subcaminho, informe o base path no build:

```bash
VITE_BASE_PATH=/nome-do-repositorio/ npm run build
```

O diretório gerado para publicação é `landing/dist`.

## Páginas públicas

A landing expõe:

- `/` — landing pública com proposta do Palpite!, FAQ e formulário Beta Android
- `/privacy` — política de privacidade
- `/terms` — termos de uso
- `/account-deletion` — fluxo de exclusão de conta e dados para uso no Google Play Console

A PWA autenticada é gerada pelo app Expo em `frontend` via `npm run build:web`.

## Cadastro Beta Android

O formulário Android salva cadastros temporariamente em `localStorage`. A função `registerTester` em `src/services/testerRegistration.ts` isola o ponto de integração futura com Supabase, Google Groups ou Play Store Closed Testing.
