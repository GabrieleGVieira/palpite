const baseUrl = import.meta.env.BASE_URL;

export function assetPath(path: string) {
  return `${baseUrl}${path.replace(/^\//, '')}`;
}

export function appPath(path = '') {
  return `${baseUrl}${path.replace(/^\//, '')}`;
}

export function currentAppPath() {
  const basePath = baseUrl.replace(/\/$/, '');
  const path = window.location.pathname.replace(/\/$/, '') || '/';

  if (!basePath) {
    return path;
  }

  return path.replace(basePath, '') || '/';
}
