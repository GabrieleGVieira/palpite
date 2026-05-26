import { StyleSheet, Text, View } from 'react-native';

import type { MatchPrediction } from '../types/prediction';

type Props = {
  explanation?: MatchPrediction['explanation'];
};

export function AiExplanationBox({ explanation }: Props) {
  if (!explanation) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.summary}>{explanation.summary}</Text>

      {explanation.main_reasons.map((reason) => (
        <View key={reason} style={styles.reasonRow}>
          <Text style={styles.bullet}>•</Text>
          <Text style={styles.reason}>{reason}</Text>
        </View>
      ))}

      {explanation.risk_alert ? (
        <Text style={styles.note}>Atenção: {explanation.risk_alert}</Text>
      ) : null}

      {explanation.user_tip ? <Text style={styles.note}>{explanation.user_tip}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  bullet: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '900',
    width: 12,
  },
  container: {
    backgroundColor: '#f5f8ef',
    borderColor: '#dfeadd',
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 12,
  },
  note: {
    color: '#486654',
    fontSize: 12,
    fontWeight: '700',
    lineHeight: 17,
  },
  reason: {
    color: '#123d2a',
    flex: 1,
    fontSize: 12,
    fontWeight: '700',
    lineHeight: 17,
  },
  reasonRow: {
    flexDirection: 'row',
    gap: 4,
  },
  summary: {
    color: '#123d2a',
    fontSize: 13,
    fontWeight: '800',
    lineHeight: 18,
  },
});
