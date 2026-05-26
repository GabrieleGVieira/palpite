import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import type { MatchPrediction, ScoreSuggestion } from '../types/prediction';
import { AiExplanationBox } from './AiExplanationBox';
import { ExpectedGoalsRow } from './ExpectedGoalsRow';
import { ProbabilityBar } from './ProbabilityBar';
import { TopScoresList } from './TopScoresList';

type MatchInfo = {
  away_team: string;
  home_team: string;
};

type Props = {
  error?: unknown;
  isLoading: boolean;
  match: MatchInfo;
  onUseSuggestion?: (score: ScoreSuggestion) => void;
  prediction?: MatchPrediction | null;
};

export function MatchPredictionCard({
  error,
  isLoading,
  match,
  onUseSuggestion,
  prediction,
}: Props) {
  const [isOpen, setIsOpen] = useState(false);

  if (isLoading) {
    return (
      <View style={styles.card}>
        <CardHeader isOpen={false} />
        <Text style={styles.summaryText}>Carregando previsão...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.card}>
        <CardHeader isOpen={false} />
        <Text style={styles.summaryText}>Não foi possível carregar a previsão agora.</Text>
      </View>
    );
  }

  if (!prediction) {
    return (
      <View style={styles.card}>
        <View style={styles.summaryButton}>
          <CardHeader isOpen={false} showChevron={false} />
          <Text style={styles.summaryText}>Previsão ainda não disponível para este jogo.</Text>
        </View>
      </View>
    );
  }

  const hasGoals =
    typeof prediction.goals?.expected_home_goals === 'number' &&
    typeof prediction.goals?.expected_away_goals === 'number';
  const hasTopScores = Boolean(prediction.top_scores?.length);
  const hasExplanation = Boolean(prediction.explanation);
  const bestScore = prediction.top_scores?.[0];
  const favorite = predictionFavorite(prediction, match);

  return (
    <View style={styles.card}>
      <Pressable
        accessibilityRole="button"
        onPress={() => setIsOpen((current) => !current)}
        style={({ pressed }) => [styles.summaryButton, pressed && styles.summaryButtonPressed]}>
        <CardHeader isOpen={isOpen} />
        <View style={styles.summaryRow}>
          <Text style={styles.summaryText} numberOfLines={1}>
            {bestScore ? `Sugestão: ${bestScore.score}` : favorite.label}
          </Text>
          <Text style={styles.summaryMeta}>
            {formatPercent(bestScore?.probability ?? favorite.probability)}
          </Text>
        </View>
      </Pressable>

      {isOpen ? (
        <View style={styles.content}>
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Probabilidades</Text>
            <ProbabilityBar
              label={`Vitória ${match.home_team}`}
              probability={prediction.probabilities.home_win}
            />
            <ProbabilityBar label="Empate" probability={prediction.probabilities.draw} />
            <ProbabilityBar
              label={`Vitória ${match.away_team}`}
              probability={prediction.probabilities.away_win}
            />
          </View>

          {hasGoals ? (
            <ExpectedGoalsRow
              awayTeam={match.away_team}
              expectedAwayGoals={prediction.goals?.expected_away_goals}
              expectedHomeGoals={prediction.goals?.expected_home_goals}
              homeTeam={match.home_team}
            />
          ) : null}

          {hasTopScores ? (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Escolha um placar</Text>
              <TopScoresList onSelectScore={onUseSuggestion} scores={prediction.top_scores} />
            </View>
          ) : null}

          {hasExplanation ? (
            <View style={styles.section}>
              <Text style={styles.sectionTitle}>Análise da IA</Text>
              <AiExplanationBox explanation={prediction.explanation} />
            </View>
          ) : null}

          <Text style={styles.disclaimer}>
            Essa previsão é uma estimativa baseada em dados. Use como apoio para o seu palpite.
          </Text>
        </View>
      ) : null}
    </View>
  );
}

function CardHeader({ isOpen, showChevron = true }: { isOpen: boolean; showChevron?: boolean }) {
  return (
    <View style={styles.header}>
      <Text style={styles.title}>PalpitAI para este jogo</Text>
      {showChevron && <Text style={styles.chevron}>{isOpen ? '▲' : '▼'}</Text>}
    </View>
  );
}

function predictionFavorite(prediction: MatchPrediction, match: MatchInfo) {
  const options = [
    { label: `Vitória ${match.home_team}`, probability: prediction.probabilities.home_win },
    { label: 'Empate', probability: prediction.probabilities.draw },
    { label: `Vitória ${match.away_team}`, probability: prediction.probabilities.away_win },
  ];

  return options.sort((a, b) => b.probability - a.probability)[0];
}

function formatPercent(probability: number) {
  return `${Math.round(probability)}%`;
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#f9fbf4',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    marginTop: 12,
    marginBottom: 12,
    overflow: 'hidden',
    padding: 8,
  },
  chevron: {
    color: '#1f7a4a',
    fontSize: 17,
    fontWeight: '900',
    lineHeight: 20,
    minWidth: 18,
    textAlign: 'center',
  },
  content: {
    borderTopColor: '#dfeadd',
    borderTopWidth: 1,
    gap: 10,
    padding: 12,
  },
  disclaimer: {
    color: '#486654',
    fontSize: 11,
    fontWeight: '700',
    lineHeight: 16,
  },
  header: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 8,
    justifyContent: 'space-between',
  },
  section: {
    gap: 8,
  },
  sectionTitle: {
    color: '#1f7a4a',
    fontSize: 12,
    fontWeight: '900',
    textTransform: 'uppercase',
  },
  stateText: {
    color: '#486654',
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18,
  },
  summaryButton: {
    gap: 6,
    padding: 12,
  },
  summaryButtonPressed: {
    backgroundColor: '#f5f8ef',
  },
  summaryMeta: {
    color: '#1f7a4a',
    fontSize: 12,
    fontWeight: '900',
  },
  summaryRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 10,
    justifyContent: 'space-between',
  },
  summaryText: {
    color: '#486654',
    flex: 1,
    fontSize: 12,
    fontWeight: '800',
    lineHeight: 16,
  },
  title: {
    color: '#123d2a',
    fontSize: 15,
    fontWeight: '900',
  },
});
