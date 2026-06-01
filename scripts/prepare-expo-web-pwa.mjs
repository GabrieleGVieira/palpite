import { copyFile, readdir, readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

const outputDir = process.argv[2] ?? 'frontend/dist';
const publicDir = process.argv[3] ?? (outputDir.includes('frontend') ? 'frontend/public' : 'public');
const basePath = normalizeBasePath(process.env.EXPO_WEB_BASE_PATH ?? '/');
const manifestPath = join(outputDir, 'manifest.json');
const serviceWorkerPath = join(outputDir, 'sw.js');
const indexPath = join(outputDir, 'index.html');

await copyFile(join(publicDir, 'favicon.png'), join(outputDir, 'favicon.png'));
await copyFile(join(publicDir, 'logo192.png'), join(outputDir, 'logo192.png'));
await copyFile(join(publicDir, 'logo512.png'), join(outputDir, 'logo512.png'));

const manifest = {
  name: 'Palpite!',
  short_name: 'Palpite!',
  description: 'Bolões inteligentes da Copa com IA.',
  start_url: withBasePath(basePath, ''),
  display: 'standalone',
  background_color: '#4A0F1B',
  theme_color: '#4A0F1B',
  orientation: 'portrait',
  icons: [
    {
      src: withBasePath(basePath, 'logo192.png'),
      sizes: '192x192',
      type: 'image/png',
    },
    {
      src: withBasePath(basePath, 'logo512.png'),
      sizes: '512x512',
      type: 'image/png',
    },
  ],
};

await writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

const serviceWorker = `
const CACHE_NAME = 'palpitai-pwa-v1';
const BASE_PATH = new URL(self.registration.scope).pathname.replace(/\\/$/, '');
const CORE_ASSETS = [
  '',
  'index.html',
  'manifest.json',
  'favicon.png',
  'logo192.png',
  'logo512.png',
].map((asset) => \`\${BASE_PATH}/\${asset}\`.replace(/\\/$/, '/'));

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) =>
      Promise.all(CORE_ASSETS.map((asset) => cache.add(asset).catch(() => undefined))),
    ),
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key)))),
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  const request = event.request;

  if (request.method !== 'GET' || new URL(request.url).origin !== self.location.origin) {
    return;
  }

  if (request.mode === 'navigate') {
    event.respondWith(fetch(request).catch(() => caches.match(\`\${BASE_PATH}/\`.replace(/\\/$/, '/'))));
    return;
  }

  event.respondWith(
    caches.match(request).then((cached) => {
      if (cached) return cached;

      return fetch(request).then((response) => {
        if (response.ok && ['style', 'script', 'image', 'font'].includes(request.destination)) {
          const copy = response.clone();
          caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
        }
        return response;
      });
    }),
  );
});
`.trimStart();

await writeFile(serviceWorkerPath, serviceWorker);

let html = await readFile(indexPath, 'utf8');
html = html
  .replaceAll('href="/favicon.ico"', `href="${basePath}favicon.ico"`)
  .replaceAll('src="/_expo/', `src="${basePath}_expo/`)
  .replaceAll('href="/_expo/', `href="${basePath}_expo/`);

const pwaHead = `
    <meta name="theme-color" content="#4A0F1B" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-title" content="Palpite!" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <link rel="apple-touch-icon" href="${basePath}logo192.png" />
    <link rel="manifest" href="${basePath}manifest.json" />`;

if (!html.includes('rel="manifest"')) {
  html = html.replace('</head>', `${pwaHead}\n  </head>`);
}

if (!html.includes('navigator.serviceWorker.register')) {
  html = html.replace(
    '</body>',
    `  <script>
    if ('serviceWorker' in navigator) {
      window.addEventListener('load', function () {
        navigator.serviceWorker.register('${basePath}sw.js', { scope: '${basePath}' }).catch(function () {});
      });
    }
  </script>
</body>`,
  );
}

await writeFile(indexPath, html);
await prefixBundleAssetPaths(join(outputDir, '_expo'), basePath);

function normalizeBasePath(value) {
  const withLeadingSlash = value.startsWith('/') ? value : `/${value}`;
  return withLeadingSlash.endsWith('/') ? withLeadingSlash : `${withLeadingSlash}/`;
}

function withBasePath(basePath, asset) {
  return `${basePath}${asset}`;
}

async function prefixBundleAssetPaths(dir, basePath) {
  let entries;

  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return;
  }

  await Promise.all(
    entries.map(async (entry) => {
      const entryPath = join(dir, entry.name);

      if (entry.isDirectory()) {
        await prefixBundleAssetPaths(entryPath, basePath);
        return;
      }

      if (!entry.isFile() || !entry.name.endsWith('.js')) {
        return;
      }

      const source = await readFile(entryPath, 'utf8');
      const updated = source.replaceAll('"/assets/', `"${basePath}assets/`);

      if (updated !== source) {
        await writeFile(entryPath, updated);
      }
    }),
  );
}
