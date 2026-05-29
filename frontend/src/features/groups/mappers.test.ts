import { describe, expect, it } from 'vitest';

import type { Group, GroupMatch, RankingEntry } from './types';
import {
  sortGroupsByPendingRequests,
  sortMatchesByKickoff,
  sortRankingByPosition,
} from './mappers';

describe('group mappers', () => {
  it('sorts groups by pending requests and name', () => {
    const groups = [
      group({ name: 'Zebra', pending_requests_count: 0 }),
      group({ name: 'Brasil', pending_requests_count: 2 }),
      group({ name: 'Argentina', pending_requests_count: 2 }),
    ];

    expect(sortGroupsByPendingRequests(groups).map((nextGroup) => nextGroup.name)).toEqual([
      'Argentina',
      'Brasil',
      'Zebra',
    ]);
  });

  it('sorts matches by kickoff', () => {
    const matches = [
      match({ id: '2', kickoff_at: '2026-06-13T10:00:00Z' }),
      match({ id: '1', kickoff_at: '2026-06-12T10:00:00Z' }),
    ];

    expect(sortMatchesByKickoff(matches).map((nextMatch) => nextMatch.id)).toEqual(['1', '2']);
  });

  it('sorts ranking by position', () => {
    const ranking = [
      rankingEntry({ position: 3, user_id: '3' }),
      rankingEntry({ position: 1, user_id: '1' }),
    ];

    expect(sortRankingByPosition(ranking).map((entry) => entry.user_id)).toEqual(['1', '3']);
  });
});

function group(override: Partial<Group>): Group {
  return {
    block_pending_predictions: false,
    created_at: '',
    description: '',
    id: '',
    invite_code: '',
    is_paid: false,
    is_private: true,
    match_scope: 'all',
    member_count: 1,
    name: '',
    owner_id: '',
    participant_limit: null,
    payment_amount: 0,
    pending_requests_count: 0,
    role: 'member',
    selected_teams: [],
    status: 'active',
    ...override,
  };
}

function match(override: Partial<GroupMatch>): GroupMatch {
  return {
    away_team: '',
    final_away_score: null,
    final_home_score: null,
    finished_at: null,
    home_team: '',
    id: '',
    kickoff_at: '',
    my_prediction: null,
    stage: '',
    status: 'scheduled',
    ...override,
  };
}

function rankingEntry(override: Partial<RankingEntry>): RankingEntry {
  return {
    display_name: '',
    position: 0,
    total_points: 0,
    user_id: '',
    ...override,
  };
}
