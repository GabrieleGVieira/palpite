# Palpite! Web/PWA

## Desenvolvimento local

```sh
npm run web
```

Esse comando executa o Expo Web com `npx expo start --web`.

## Build

```sh
npx expo export --platform web
```

O script equivalente do projeto também está disponível:

```sh
npm run build:web
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
- O PWA usa `public/manifest.json`, `public/sw.js`, `public/logo192.png`, `public/logo512.png` e `public/favicon.png`.
- O upload de avatar usa `expo-file-system` apenas no Android/iOS e usa `fetch(uri).blob()` no Web.
- A seleção de imagem mantém `expo-image-picker`; no Web a solicitação explícita de permissão da galeria é ignorada.
- `AsyncStorage`, `expo-clipboard`, `expo-status-bar` e `react-native-safe-area-context` foram mantidos por terem suporte ou fallback compatível no Expo Web.
