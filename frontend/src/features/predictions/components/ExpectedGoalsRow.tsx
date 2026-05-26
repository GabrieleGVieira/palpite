import { StyleSheet, Text, View } from 'react-native';

type Props = {
  awayTeam: string;
  expectedAwayGoals?: number;
  expectedHomeGoals?: number;
  homeTeam: string;
};

export function ExpectedGoalsRow({
  awayTeam,
  expectedAwayGoals,
  expectedHomeGoals,
  homeTeam,
}: Props) {
  const hasExpectedGoals =
    typeof expectedHomeGoals === 'number' && typeof expectedAwayGoals === 'number';

  if (!hasExpectedGoals) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Gols esperados</Text>
      <Text style={styles.value}>
        {homeTeam} {expectedHomeGoals.toFixed(1)} x {expectedAwayGoals.toFixed(1)} {awayTeam}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    borderColor: '#dfeadd',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 8,
    justifyContent: 'space-between',
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  label: {
    color: '#486654',
    flexShrink: 0,
    fontSize: 11,
    fontWeight: '900',
    textTransform: 'uppercase',
  },
  value: {
    color: '#123d2a',
    flex: 1,
    fontSize: 12,
    fontWeight: '900',
    lineHeight: 16,
    textAlign: 'right',
  },
});
