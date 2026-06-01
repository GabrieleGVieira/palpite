import { ActivityIndicator, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';

import type { GroupMatch } from '../../services/groups';
import type { ScoreDraft } from '../../types';
import { formatDate, formatMatchStage, formatMatchStatus } from '../../utils/groupDetailFormatters';
import { MatchPredictionCard } from '../../../predictions/components/MatchPredictionCard';
import { useMatchPrediction } from '../../../predictions/hooks/useMatchPrediction';
import { isScheduledStatus } from '../../../predictions/utils/predictionStatus';
import { FinishButton } from '../../../../shared/components/FinishButton';

type Props = {
  draft: ScoreDraft;
  isSaving: boolean;
  match: GroupMatch;
  onChangeDraft: (matchID: string, key: keyof ScoreDraft, value: string) => void;
  onCreateChallenge: (matchID: string) => void;
  onSavePrediction: (match: GroupMatch) => Promise<void>;
};

export function GroupDetailMatchCard({
  draft,
  isSaving,
  match,
  onChangeDraft,
  onCreateChallenge,
  onSavePrediction,
}: Props) {
  const hasStarted = new Date(match.kickoff_at).getTime() <= Date.now();
  const normalizedStatus = match.status.trim().toLowerCase();
  const isLive = normalizedStatus === 'live' || normalizedStatus === 'in_play';
  const isPredictionClosed = hasStarted || !isScheduledStatus(match.status);
  const liveHomeScore = match.final_home_score ?? '-';
  const liveAwayScore = match.final_away_score ?? '-';
  const disabledLabel = getDisabledPredictionLabel(match.status, hasStarted);
  const matchPrediction = useMatchPrediction(match.id, match.status);

  function handleUseSuggestion(score: { away: number; home: number }) {
    onChangeDraft(match.id, 'homeScore', String(score.home));
    onChangeDraft(match.id, 'awayScore', String(score.away));
  }

  return (
    <View style={styles.matchCard}>
      <View style={styles.matchHeader}>
        <View>
          <Text style={styles.stage}>{formatMatchStage(match.stage)}</Text>
          <Text style={styles.matchStatus}>{formatMatchStatus(match.status)}</Text>
        </View>
        <Text style={styles.kickoff}>{formatDate(match.kickoff_at)}</Text>
      </View>

      {match.finished_at && match.final_home_score !== null && match.final_away_score !== null ? (
        <View style={styles.resultBox}>
          <Text style={styles.resultLabel}>Resultado final</Text>
          <Text style={styles.resultText}>
            {match.home_team} {match.final_home_score} x {match.final_away_score} {match.away_team}
          </Text>
        </View>
      ) : null}

      <View style={styles.scoreRow}>
        <Text style={styles.teamName}>{match.home_team}</Text>
        <TextInput
          editable={!hasStarted && !isSaving}
          keyboardType="number-pad"
          onChangeText={(value) => onChangeDraft(match.id, 'homeScore', value)}
          style={[styles.scoreInput, hasStarted && styles.inputDisabled]}
          value={draft.homeScore}
        />
        <Text style={styles.scoreSeparator}>x</Text>
        <TextInput
          editable={!hasStarted && !isSaving}
          keyboardType="number-pad"
          onChangeText={(value) => onChangeDraft(match.id, 'awayScore', value)}
          style={[styles.scoreInput, hasStarted && styles.inputDisabled]}
          value={draft.awayScore}
        />
        <Text style={styles.teamName}>{match.away_team}</Text>
      </View>

      {match.my_prediction ? (
        <View style={styles.predictionSummary}>
          <Text style={styles.predictionText}>
            Seu palpite: {match.my_prediction.home_score} x {match.my_prediction.away_score}
          </Text>
          {match.my_prediction.points !== null && normalizedStatus === 'finished' ? (
            <Text style={styles.pointsText}>{match.my_prediction.points} pts</Text>
          ) : null}
        </View>
      ) : (
        <Text style={[styles.predictionText, styles.predictionTextSolo]}>
          Você ainda não deu seu palpite neste jogo.
        </Text>
      )}

      {matchPrediction.shouldShowPrediction ? (
        <MatchPredictionCard
          error={matchPrediction.error}
          isLoading={matchPrediction.isLoading}
          match={match}
          onUseSuggestion={handleUseSuggestion}
          prediction={matchPrediction.data}
        />
      ) : null}

      {isLive ? (
        <View style={styles.liveBox}>
          <View style={styles.liveLabelRow}>
            <ActivityIndicator color="#a03222" size="small" />
            <Text style={styles.liveLabel}>Ao vivo</Text>
          </View>
          <Text style={styles.liveScore}>
            {match.home_team} {liveHomeScore} x {liveAwayScore} {match.away_team}
          </Text>
        </View>
      ) : (
        <View style={styles.actionStack}>
          <FinishButton
            disabledLabel={disabledLabel}
            isDisabled={isPredictionClosed}
            isLoading={isSaving}
            onPress={() => onSavePrediction(match)}
            loadingLabel="Salvando..."
            waitingLabel="Salvar palpite"
          />
          {!isPredictionClosed ? (
            <Pressable onPress={() => onCreateChallenge(match.id)} style={styles.challengeButton}>
              <Text style={styles.challengeButtonText}>Desafiar amigo</Text>
            </Pressable>
          ) : null}
        </View>
      )}
    </View>
  );
}

function getDisabledPredictionLabel(status: GroupMatch['status'], hasStarted: boolean) {
  const normalizedStatus = status.trim().toLowerCase();

  if (normalizedStatus === 'finished') {
    return 'Jogo encerrado';
  }

  if (normalizedStatus === 'cancelled') {
    return 'Jogo cancelado';
  }

  if (normalizedStatus === 'postponed') {
    return 'Palpite indisponível';
  }

  if (hasStarted) {
    return 'Palpites encerrados';
  }

  return 'Salvar palpite';
}

const styles = StyleSheet.create({
  matchCard: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 16,
  },
  matchHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  stage: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '800',
  },
  matchStatus: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '700',
    marginTop: 3,
  },
  kickoff: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '700',
  },
  resultBox: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    marginTop: 14,
    padding: 12,
  },
  resultLabel: {
    color: '#486654',
    fontSize: 11,
    fontWeight: '800',
    textTransform: 'uppercase',
  },
  resultText: {
    color: '#123d2a',
    fontSize: 14,
    fontWeight: '800',
    marginTop: 4,
  },
  scoreRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 8,
    marginTop: 16,
  },
  teamName: {
    color: '#123d2a',
    flex: 1,
    fontSize: 14,
    fontWeight: '800',
  },
  scoreInput: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 18,
    fontWeight: '800',
    minHeight: 46,
    textAlign: 'center',
    width: 46,
  },
  inputDisabled: {
    backgroundColor: '#edf3e8',
    color: '#7c8898',
  },
  scoreSeparator: {
    color: '#486654',
    fontSize: 16,
    fontWeight: '800',
  },
  predictionText: {
    color: '#486654',
    flex: 1,
    fontSize: 13,
    lineHeight: 18,
  },
  predictionSummary: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 12,
    marginTop: 12,
  },
  predictionTextSolo: {
    marginTop: 12,
  },
  pointsText: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '900',
    overflow: 'hidden',
    paddingHorizontal: 10,
    paddingVertical: 6,
    marginBottom: 5,
  },
  liveBox: {
    alignItems: 'center',
    backgroundColor: '#fff2f0',
    borderColor: '#f0b8b1',
    borderRadius: 8,
    borderWidth: 1,
    gap: 6,
    justifyContent: 'center',
    marginTop: 14,
    minHeight: 52,
    paddingHorizontal: 12,
    paddingVertical: 10,
  },
  liveLabel: {
    color: '#a03222',
    fontSize: 12,
    fontWeight: '900',
    textTransform: 'uppercase',
  },
  liveLabelRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 6,
  },
  liveScore: {
    color: '#123d2a',
    fontSize: 14,
    fontWeight: '900',
    textAlign: 'center',
  },
  actionStack: {
    gap: 10,
    marginTop: 14,
  },
  challengeButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderRadius: 8,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 46,
  },
  challengeButtonText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '900',
  },
  saveButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    marginTop: 14,
    minHeight: 48,
  },
  buttonDisabled: {
    opacity: 0.72,
  },
  saveButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '800',
  },
});
