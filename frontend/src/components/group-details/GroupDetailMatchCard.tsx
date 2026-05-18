import { StyleSheet, Text, TextInput, View } from 'react-native';

import type { GroupMatch } from '../../services/groups';
import type { ScoreDraft } from '../../types/groupDetail';
import { formatDate, formatMatchStage, formatMatchStatus } from '../../utils/groupDetailFormatters';
import { FinishButton } from '../global/FinishButton';

type Props = {
  draft: ScoreDraft;
  isSaving: boolean;
  match: GroupMatch;
  onChangeDraft: (matchID: string, key: keyof ScoreDraft, value: string) => void;
  onSavePrediction: (match: GroupMatch) => Promise<void>;
};

export function GroupDetailMatchCard({
  draft,
  isSaving,
  match,
  onChangeDraft,
  onSavePrediction,
}: Props) {
  const hasStarted = new Date(match.kickoff_at).getTime() <= Date.now();
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
          {match.my_prediction.points !== null ? (
            <Text style={styles.pointsText}>{match.my_prediction.points} pts</Text>
          ) : null}
        </View>
      ) : (
        <Text style={[styles.predictionText, styles.predictionTextSolo]}>
          Você ainda não palpitou neste jogo.
        </Text>
      )}

      <FinishButton
        isLoading={hasStarted || isSaving}
        onPress={() => onSavePrediction(match)}
        loadingLabel="Salvando..."
        waitingLabel="Salvar palpite"
      />
    </View>
  );
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
