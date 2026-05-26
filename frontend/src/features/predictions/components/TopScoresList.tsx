import { Pressable, StyleSheet, Text, View } from 'react-native';

import type { PredictionScore, ScoreSuggestion } from '../types/prediction';
import { parseScoreSuggestion, topThreeScores } from '../utils/scoreSuggestion';

type Props = {
  onSelectScore?: (score: ScoreSuggestion) => void;
  scores?: PredictionScore[];
};

export function TopScoresList({ onSelectScore, scores }: Props) {
  const visibleScores = topThreeScores(scores);

  if (visibleScores.length === 0) {
    return null;
  }

  function handleSelect(score: PredictionScore) {
    const suggestion = parseScoreSuggestion(score.score);
    if (suggestion) {
      onSelectScore?.(suggestion);
    }
  }

  return (
    <View style={styles.container}>
      {visibleScores.map((score, index) => (
        <ScoreOption key={score.score} score={score} rank={index + 1} onSelect={handleSelect} />
      ))}
    </View>
  );
}

type ScoreOptionProps = {
  onSelect: (score: PredictionScore) => void;
  rank: number;
  score: PredictionScore;
};

function ScoreOption({ onSelect, rank, score }: ScoreOptionProps) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={() => onSelect(score)}
      style={({ pressed }) => [styles.scoreRow, pressed && styles.scoreRowPressed]}
    >
      <Text style={styles.rank}>{rank}</Text>
      <Text style={styles.score}>{score.score}</Text>
      <Text style={styles.probability}>{formatProbability(score.probability)}</Text>
      <Text style={styles.useLabel}>Escolher</Text>
    </Pressable>
  );
}

function formatProbability(probability: number) {
  return `${Math.round(probability)}%`;
}

const styles = StyleSheet.create({
  container: {
    gap: 6,
  },
  probability: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '900',
    minWidth: 42,
    textAlign: 'right',
  },
  rank: {
    color: '#78917f',
    fontSize: 12,
    fontWeight: '900',
    width: 18,
  },
  score: {
    color: '#123d2a',
    flex: 1,
    fontSize: 13,
    fontWeight: '900',
  },
  scoreRow: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderBottomColor: '#edf3e8',
    borderColor: '#dfeadd',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 8,
    minHeight: 36,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  scoreRowPressed: {
    backgroundColor: '#f5f8ef',
  },
  useLabel: {
    color: '#1f7a4a',
    fontSize: 11,
    fontWeight: '900',
  },
});
