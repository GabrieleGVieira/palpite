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

O formulário Android envia `POST /api/beta/android` para o backend configurado em `VITE_API_URL`. O backend salva o e-mail, adiciona o usuário ao Google Group vinculado à track de teste no Play Console e retorna a URL configurada em `PLAY_STORE_BETA_URL` para redirecionamento.

Configure `VITE_API_URL` apontando para a API pública:

```env
VITE_API_URL=https://palpitai-api.onrender.com
```

Não configure chaves Google na landing. `GOOGLE_PRIVATE_KEY`, `GOOGLE_SERVICE_ACCOUNT_EMAIL`, `GOOGLE_GROUP_EMAIL` e `PLAY_STORE_BETA_URL` pertencem somente ao backend.
