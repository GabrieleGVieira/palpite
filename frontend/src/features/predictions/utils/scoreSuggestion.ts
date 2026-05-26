import type { ScoreSuggestion } from '../types/prediction';

export function parseScoreSuggestion(score?: string): ScoreSuggestion | null {
  if (!score) {
    return null;
  }

  const match = score.trim().match(/^(\d{1,2})\s*x\s*(\d{1,2})$/i);

  if (!match) {
    return null;
  }

  return {
    away: Number(match[2]),
    home: Number(match[1]),
  };
}

export function topThreeScores<T>(scores?: T[]): T[] {
  return Array.isArray(scores) ? scores.slice(0, 3) : [];
}
