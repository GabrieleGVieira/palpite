import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useMatchPrediction } from './useMatchPrediction';

const { useQueryMock } = vi.hoisted(() => ({
  useQueryMock: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: useQueryMock,
}));

vi.mock('../api/getMatchPrediction', () => ({
  getMatchPrediction: vi.fn(),
}));

describe('useMatchPrediction', () => {
  beforeEach(() => {
    useQueryMock.mockReset();
    useQueryMock.mockReturnValue({
      data: null,
      error: null,
      isLoading: false,
    });
  });

  it('disables the request when the match is not scheduled', () => {
    const result = useMatchPrediction('match-1', 'FINISHED');
    const options = useQueryMock.mock.calls[0][0];

    expect(options.enabled).toBe(false);
    expect(result.shouldShowPrediction).toBe(false);
  });

  it('enables the request when the match is scheduled', () => {
    const result = useMatchPrediction('match-1', 'scheduled');
    const options = useQueryMock.mock.calls[0][0];

    expect(options.enabled).toBe(true);
    expect(options.queryKey).toEqual(['matchPrediction', 'match-1']);
    expect(result.shouldShowPrediction).toBe(true);
  });
});
