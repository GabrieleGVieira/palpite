export function websocketURL(value: string) {
  if (value.startsWith('https://')) {
    return value.replace('https://', 'wss://');
  }

  return value.replace('http://', 'ws://');
}
