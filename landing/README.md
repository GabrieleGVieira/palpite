# PalpitAI Landing

Landing page independente do app mobile, criada com React, TypeScript, Vite e CSS simples.

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

Por padrão o Vite usa `base: "/"`, adequado para domínio próprio como `https://palpitai.com.br`.

Se o deploy for em um repositório GitHub Pages com subcaminho, informe o base path no build:

```bash
VITE_BASE_PATH=/nome-do-repositorio/ npm run build
```

O diretório gerado para publicação é `landing/dist`.

## Integrações futuras

O formulário Android salva cadastros temporariamente em `localStorage`. A função
`registerTester` em `src/services/testerRegistration.ts` já isola o ponto para integrar depois com
Supabase, Google Groups e Play Store Closed Testing.
