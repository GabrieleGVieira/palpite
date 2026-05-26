import { describe, expect, it } from 'vitest';

import { parseScoreSuggestion, topThreeScores } from './scoreSuggestion';

describe('parseScoreSuggestion', () => {
  it('parses a likely score into home and away values', () => {
    expect(parseScoreSuggestion('2x1')).toEqual({ away: 1, home: 2 });
    expect(parseScoreSuggestion(' 10 x 0 ')).toEqual({ away: 0, home: 10 });
  });

  it('returns null for missing or invalid scores', () => {
    expect(parseScoreSuggestion()).toBeNull();
    expect(parseScoreSuggestion('2-1')).toBeNull();
  });
});

describe('topThreeScores', () => {
  it('returns only the first three scores', () => {
    expect(topThreeScores(['1x0', '1x1', '2x1', '2x0'])).toEqual(['1x0', '1x1', '2x1']);
  });
});
