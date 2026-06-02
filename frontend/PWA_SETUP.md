# Palpite! Web/PWA

## Desenvolvimento local

```sh
npm run web
```

Esse comando executa o Expo Web com `npx expo start --web`.

## Build

```sh
npm run build:web
```

Esse script executa o export web do Expo e depois prepara os arquivos PWA em `dist/`.

Para rodar apenas o export do Expo:

```sh
npm run export:web
```

## Deploy

```sh
npx eas deploy
```

## Instalação no iPhone

1. Abrir a URL publicada no Safari.
2. Tocar em Compartilhar.
3. Tocar em Adicionar à Tela Inicial.
4. Confirmar o nome Palpite!.

## Notas de compatibilidade Web

- O slug Expo permanece `palpitai`.
- O PWA publicado usa os arquivos preparados em `dist/` pelo script `scripts/prepare-expo-web-pwa.mjs`.
- O upload de avatar usa `expo-file-system` apenas no Android/iOS e usa `fetch(uri).blob()` no Web.
- A seleção de imagem mantém `expo-image-picker`; no Web a solicitação explícita de permissão da galeria é ignorada.
- `AsyncStorage`, `expo-clipboard`, `expo-status-bar` e `react-native-safe-area-context` foram mantidos por terem suporte ou fallback compatível no Expo Web.
