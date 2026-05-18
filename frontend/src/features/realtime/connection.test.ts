import { describe, expect, it } from 'vitest';

import { websocketURL } from './url';

describe('websocketURL', () => {
  it('converts http API URLs to ws', () => {
    expect(websocketURL('http://localhost:3000')).toBe('ws://localhost:3000');
  });

  it('converts https API URLs to wss', () => {
    expect(websocketURL('https://api.example.com')).toBe('wss://api.example.com');
  });
});
