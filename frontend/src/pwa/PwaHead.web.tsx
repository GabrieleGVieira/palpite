import { useEffect } from 'react';

const themeColor = '#f5f8ef';

export function PwaHead() {
  useEffect(() => {
    upsertMeta('theme-color', themeColor);
    upsertMeta('apple-mobile-web-app-capable', 'yes');
    upsertMeta('apple-mobile-web-app-status-bar-style', 'black-translucent');
    upsertMeta('apple-mobile-web-app-title', 'Palpite!');
    upsertLink('manifest', '/manifest.json');
    upsertLink('apple-touch-icon', '/logo192.png');
    upsertFavicon('/favicon.png');
    registerServiceWorker();
  }, []);

  return null;
}

function upsertMeta(name: string, content: string) {
  let element = document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`);

  if (!element) {
    element = document.createElement('meta');
    element.name = name;
    document.head.appendChild(element);
  }

  element.content = content;
}

function upsertLink(rel: string, href: string) {
  let element = document.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`);

  if (!element) {
    element = document.createElement('link');
    element.rel = rel;
    document.head.appendChild(element);
  }

  element.href = href;
}

function upsertFavicon(href: string) {
  const icon =
    document.querySelector<HTMLLinkElement>('link[rel="icon"]') ??
    document.querySelector<HTMLLinkElement>('link[rel="shortcut icon"]');

  if (icon) {
    icon.href = href;
    return;
  }

  upsertLink('icon', href);
}

function registerServiceWorker() {
  if (!('serviceWorker' in navigator) || window.location.protocol !== 'https:') {
    return;
  }

  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch(() => {});
  });
}
