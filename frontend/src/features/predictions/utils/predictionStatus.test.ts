import { describe, expect, it } from 'vitest';

import { isScheduledStatus } from './predictionStatus';

describe('isScheduledStatus', () => {
  it('returns true for scheduled statuses', () => {
    expect(isScheduledStatus('scheduled')).toBe(true);
    expect(isScheduledStatus('SCHEDULED')).toBe(true);
    expect(isScheduledStatus('timed')).toBe(true);
    expect(isScheduledStatus('TIMED')).toBe(true);
  });

  it('returns false for live, finished and unavailable statuses', () => {
    expect(isScheduledStatus('LIVE')).toBe(false);
    expect(isScheduledStatus('IN_PLAY')).toBe(false);
    expect(isScheduledStatus('PAUSED')).toBe(false);
    expect(isScheduledStatus('FINISHED')).toBe(false);
    expect(isScheduledStatus('POSTPONED')).toBe(false);
    expect(isScheduledStatus('CANCELLED')).toBe(false);
    expect(isScheduledStatus('SUSPENDED')).toBe(false);
    expect(isScheduledStatus(undefined)).toBe(false);
  });
});
