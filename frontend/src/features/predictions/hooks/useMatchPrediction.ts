import { useQuery } from '@tanstack/react-query';

import { getMatchPrediction } from '../api/getMatchPrediction';
import { isScheduledStatus } from '../utils/predictionStatus';

export function useMatchPrediction(matchId: string | undefined, status: string | undefined) {
  const shouldShowPrediction = Boolean(matchId) && isScheduledStatus(status);
  const predictionQuery = useQuery({
    enabled: shouldShowPrediction,
    queryFn: () => getMatchPrediction(matchId as string),
    queryKey: ['matchPrediction', matchId],
    retry: 1,
    staleTime: 5 * 60 * 1000,
  });

  return {
    data: predictionQuery.data ?? null,
    error: predictionQuery.error,
    isLoading: predictionQuery.isLoading,
    shouldShowPrediction,
  };
}
