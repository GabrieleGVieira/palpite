import { StyleSheet, Text, View } from 'react-native';

type ScoreCardProps = {
  totalPoints: number;
  isLoading: boolean;
};

export function ScoreCard({ totalPoints, isLoading }: ScoreCardProps) {
  return (
    <View style={styles.card}>
      <View>
        <Text style={styles.label}>Pontuação geral</Text>
        <Text style={styles.hint}>Somando todos os grupos ativos</Text>
      </View>
      <Text style={styles.value}>{isLoading ? '...' : totalPoints}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    alignItems: 'center',
    backgroundColor: '#123d2a',
    borderColor: '#296943',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    padding: 18,
  },
  label: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '800',
  },
  hint: {
    color: '#cde4c9',
    fontSize: 12,
    marginTop: 4,
  },
  value: {
    color: '#ffffff',
    fontSize: 34,
    fontWeight: '900',
  },
});
