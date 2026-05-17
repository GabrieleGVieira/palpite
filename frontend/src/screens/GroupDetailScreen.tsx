import { StatusBar } from 'expo-status-bar';
import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { listGroupMatches, savePrediction, type Group, type GroupMatch } from '../services/groups';

type GroupDetailScreenProps = {
  group: Group;
  onBack: () => void;
  onOpenAdmin: () => void;
};

type ScoreDraft = {
  awayScore: string;
  homeScore: string;
};

export function GroupDetailScreen({ group, onBack, onOpenAdmin }: GroupDetailScreenProps) {
  const [matches, setMatches] = useState<GroupMatch[]>([]);
  const [drafts, setDrafts] = useState<Record<string, ScoreDraft>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [savingMatchID, setSavingMatchID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const loadMatches = useCallback(async () => {
    setError(null);
    setIsLoading(true);

    try {
      const nextMatches = await listGroupMatches(group.id);
      setMatches(nextMatches);
      setDrafts(buildDrafts(nextMatches));
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Nao foi possivel carregar jogos.');
    } finally {
      setIsLoading(false);
    }
  }, [group.id]);

  useEffect(() => {
    loadMatches();
  }, [loadMatches]);

  function updateDraft(matchID: string, key: keyof ScoreDraft, value: string) {
    setDrafts((currentDrafts) => ({
      ...currentDrafts,
      [matchID]: {
        ...(currentDrafts[matchID] ?? { awayScore: '', homeScore: '' }),
        [key]: value.replace(/\D/g, '').slice(0, 2),
      },
    }));
  }

  async function handleSavePrediction(match: GroupMatch) {
    const draft = drafts[match.id];
    setError(null);
    setSuccessMessage(null);

    if (!draft?.homeScore || !draft.awayScore) {
      setError('Informe os dois placares para salvar o palpite.');
      return;
    }

    setSavingMatchID(match.id);

    try {
      const prediction = await savePrediction(group.id, match.id, {
        away_score: Number(draft.awayScore),
        home_score: Number(draft.homeScore),
      });

      setMatches((currentMatches) =>
        currentMatches.map((currentMatch) =>
          currentMatch.id === match.id
            ? {
                ...currentMatch,
                my_prediction: prediction,
              }
            : currentMatch,
        ),
      );
      setSuccessMessage('Palpite salvo.');
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : 'Nao foi possivel salvar o palpite.',
      );
    } finally {
      setSavingMatchID(null);
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <View style={styles.topBar}>
          <Pressable accessibilityLabel="Voltar" onPress={onBack} style={styles.backButton}>
            <Text style={styles.backButtonText}>‹</Text>
          </Pressable>

          {group.role === 'owner' ? (
            <Pressable onPress={onOpenAdmin} style={styles.adminButton}>
              <Text style={styles.adminButtonText}>Admin</Text>
            </Pressable>
          ) : null}
        </View>

        <View style={styles.header}>
          <Text style={styles.title}>{group.name}</Text>
          <Text style={styles.subtitle}>
            {group.match_scope === 'all'
              ? 'Todos os jogos da Copa'
              : `Selecoes: ${group.selected_teams.join(', ')}`}
          </Text>
        </View>

        {error ? <Text style={styles.errorText}>{error}</Text> : null}
        {successMessage ? <Text style={styles.successText}>{successMessage}</Text> : null}

        {isLoading ? (
          <View style={styles.loadingBox}>
            <ActivityIndicator color="#1f7a4a" />
            <Text style={styles.loadingText}>Carregando jogos...</Text>
          </View>
        ) : null}

        {!isLoading && matches.length === 0 ? (
          <View style={styles.emptyBox}>
            <Text style={styles.emptyTitle}>Nenhum jogo encontrado</Text>
            <Text style={styles.emptyText}>A lista de jogos desse bolao ainda esta vazia.</Text>
          </View>
        ) : null}

        {matches.map((match) => {
          const draft = drafts[match.id] ?? { awayScore: '', homeScore: '' };
          const hasStarted = new Date(match.kickoff_at).getTime() <= Date.now();
          const isSaving = savingMatchID === match.id;

          return (
            <View key={match.id} style={styles.matchCard}>
              <View style={styles.matchHeader}>
                <Text style={styles.stage}>{match.stage}</Text>
                <Text style={styles.kickoff}>{formatDate(match.kickoff_at)}</Text>
              </View>

              <View style={styles.scoreRow}>
                <Text style={styles.teamName}>{match.home_team}</Text>
                <TextInput
                  editable={!hasStarted && !isSaving}
                  keyboardType="number-pad"
                  onChangeText={(value) => updateDraft(match.id, 'homeScore', value)}
                  style={[styles.scoreInput, hasStarted && styles.inputDisabled]}
                  value={draft.homeScore}
                />
                <Text style={styles.scoreSeparator}>x</Text>
                <TextInput
                  editable={!hasStarted && !isSaving}
                  keyboardType="number-pad"
                  onChangeText={(value) => updateDraft(match.id, 'awayScore', value)}
                  style={[styles.scoreInput, hasStarted && styles.inputDisabled]}
                  value={draft.awayScore}
                />
                <Text style={styles.teamName}>{match.away_team}</Text>
              </View>

              {match.my_prediction ? (
                <Text style={styles.predictionText}>
                  Seu palpite: {match.my_prediction.home_score} x {match.my_prediction.away_score}
                </Text>
              ) : (
                <Text style={styles.predictionText}>Voce ainda nao palpitou neste jogo.</Text>
              )}

              <Pressable
                disabled={hasStarted || isSaving}
                onPress={() => handleSavePrediction(match)}
                style={[styles.saveButton, (hasStarted || isSaving) && styles.buttonDisabled]}>
                <Text style={styles.saveButtonText}>
                  {hasStarted ? 'Palpites encerrados' : isSaving ? 'Salvando...' : 'Salvar palpite'}
                </Text>
              </Pressable>
            </View>
          );
        })}
      </ScrollView>
    </SafeAreaView>
  );
}

function buildDrafts(matches: GroupMatch[]) {
  return Object.fromEntries(
    matches.map((match) => [
      match.id,
      {
        awayScore: match.my_prediction ? String(match.my_prediction.away_score) : '',
        homeScore: match.my_prediction ? String(match.my_prediction.home_score) : '',
      },
    ]),
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(value));
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  container: {
    backgroundColor: '#f5f8ef',
    flexGrow: 1,
    gap: 18,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  pitchMarkTop: {
    borderColor: 'rgba(255, 255, 255, 0.68)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  pitchCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 140,
    borderWidth: 2,
    height: 280,
    position: 'absolute',
    right: -128,
    top: 104,
    width: 280,
  },
  topBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
  adminButton: {
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  adminButtonText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
  },
  header: {
    paddingTop: 12,
  },
  title: {
    color: '#123d2a',
    fontSize: 34,
    fontWeight: '800',
  },
  subtitle: {
    color: '#486654',
    fontSize: 15,
    lineHeight: 22,
    marginTop: 8,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  successText: {
    color: '#1f7a4a',
    fontSize: 13,
    lineHeight: 18,
  },
  loadingBox: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 18,
  },
  loadingText: {
    color: '#486654',
    fontSize: 14,
  },
  emptyBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 18,
  },
  emptyTitle: {
    color: '#123d2a',
    fontSize: 17,
    fontWeight: '800',
  },
  emptyText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 6,
  },
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
  kickoff: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '700',
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
    fontSize: 13,
    lineHeight: 18,
    marginTop: 12,
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
