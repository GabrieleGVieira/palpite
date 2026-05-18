import { View, Text, StyleSheet } from 'react-native';
import { RankingEntry } from '../../services/groups';

export function GroupDetailRankingCard({ entry, key }: { entry: RankingEntry; key: string }) {
  return (
    <View key={key} style={styles.rankingCard}>
      <View style={styles.positionBadge}>
        <Text style={styles.positionText}>#{entry.position}</Text>
      </View>
      <View style={styles.rankingUserInfo}>
        <Text style={styles.rankingUser}>{entry.display_name || entry.user_id}</Text>
        <Text style={styles.rankingMeta}>Participante do grupo</Text>
      </View>
      <Text style={styles.rankingPoints}>{entry.total_points} pts</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  rankingCard: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 12,
    padding: 16,
  },
  positionBadge: {
    alignItems: 'center',
    backgroundColor: '#123d2a',
    borderRadius: 8,
    height: 44,
    justifyContent: 'center',
    width: 52,
  },
  positionText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '900',
  },
  rankingUserInfo: {
    flex: 1,
  },
  rankingUser: {
    color: '#123d2a',
    fontSize: 15,
    fontWeight: '800',
  },
  rankingMeta: {
    color: '#486654',
    fontSize: 12,
    marginTop: 3,
  },
  rankingPoints: {
    color: '#1f7a4a',
    fontSize: 16,
    fontWeight: '900',
  },
});
