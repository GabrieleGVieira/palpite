import { StyleSheet, Text, View } from 'react-native';

type Props = {
  label: string;
  probability: number;
};

export function ProbabilityBar({ label, probability }: Props) {
  const percent = clampPercent(probability);

  return (
    <View style={styles.container}>
      <View style={styles.labelRow}>
        <Text style={styles.label}>{label}</Text>
        <Text style={styles.value}>{formatPercent(probability)}</Text>
      </View>
      <View style={styles.track}>
        <View style={[styles.fill, { width: `${percent}%` }]} />
      </View>
    </View>
  );
}

function clampPercent(value: number) {
  if (!Number.isFinite(value)) {
    return 0;
  }

  return Math.min(Math.max(value, 0), 100);
}

function formatPercent(value: number) {
  return `${Math.round(value)}%`;
}

const styles = StyleSheet.create({
  container: {
    gap: 6,
  },
  fill: {
    backgroundColor: '#1f7a4a',
    borderRadius: 999,
    height: '100%',
  },
  label: {
    color: '#123d2a',
    flex: 1,
    fontSize: 13,
    fontWeight: '800',
  },
  labelRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 12,
    justifyContent: 'space-between',
  },
  track: {
    backgroundColor: '#dfeadd',
    borderRadius: 999,
    height: 8,
    overflow: 'hidden',
  },
  value: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '900',
  },
});
