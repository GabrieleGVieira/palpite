import { apiClient } from '../../../shared/services/apiClient';
import type { MatchPrediction } from '../types/prediction';

export async function getMatchPrediction(matchId: string): Promise<MatchPrediction | null> {
  const prediction = await apiClient<MatchPrediction | null>(`/api/v1/matches/${matchId}/prediction`, {
    fallbackError: 'Não foi possível carregar a análise da PalpitAI agora.',
    notFoundValue: null,
  });

  if (!prediction || !('match_id' in prediction)) {
    return null;
  }

  return prediction;
}
