import { ScrollView, StyleSheet, Text, View } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useQuery } from '@tanstack/react-query';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { colors, spacing } from '../../../shared/theme';
import { useAuth } from '../../auth/hooks/useAuth';
import { UserAvatar } from '../components/UserAvatar';
import { getPublicProfile } from '../services/friends';
import type { Challenge } from '../../challenges/services/challenges';

type Props = {
  onBack: () => void;
  userID: string;
};

export function PublicProfileScreen({ onBack, userID }: Props) {
  const { user } = useAuth();
  const profileQuery = useQuery({
    queryFn: () => getPublicProfile(userID),
    queryKey: ['users', userID, 'profile'],
  });

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.headerRow}>
          <BackButton onPress={onBack} />
          <Text style={styles.title}>Perfil</Text>
        </View>

        {profileQuery.isLoading ? <LoadingIndicator text="Carregando perfil..." /> : null}
        {profileQuery.error ? (
          <EmptyBox title="Perfil indisponível" text="Não foi possivel carregar este Palpiteiro." />
        ) : null}
        {profileQuery.data ? (
          <>
            <View style={styles.profileHeader}>
              <UserAvatar name={profileQuery.data.name} size={76} uri={profileQuery.data.avatarUrl} />
              <View style={styles.profileInfo}>
                <Text style={styles.name}>{profileQuery.data.name}</Text>
                <Text style={styles.memberSince}>
                  Membro desde {formatDate(profileQuery.data.joinedAt)}
                </Text>
              </View>
            </View>

            <View style={styles.statsGrid}>
              <StatCard label="Pontos" value={String(profileQuery.data.totalPoints)} />
              <StatCard label="Palpites" value={String(profileQuery.data.predictionsCount)} />
              <StatCard
                label="Ranking"
                value={profileQuery.data.globalRanking ? `#${profileQuery.data.globalRanking}` : '-'}
              />
              <StatCard label="Grupos" value={String(profileQuery.data.groupsCount)} />
            </View>

            <View style={styles.challengePanel}>
              <Text style={styles.sectionTitle}>Desafios entre vocês</Text>
              {profileQuery.data.friendshipStatus !== 'ACCEPTED' ? (
                <Text style={styles.mutedText}>Disponível apenas para amigos.</Text>
              ) : null}
              {profileQuery.data.friendshipStatus === 'ACCEPTED' && !profileQuery.data.challenges?.length ? (
                <Text style={styles.mutedText}>Nenhum desafio entre vocês ainda.</Text>
              ) : null}
              {(profileQuery.data.challenges ?? []).map((challenge) => (
                <View key={challenge.id} style={styles.challengeRow}>
                  <Text style={styles.challengeTitle}>
                    {challenge.homeTeam && challenge.awayTeam
                      ? `${challenge.homeTeam} x ${challenge.awayTeam}`
                      : 'Jogo selecionado'}
                  </Text>
                  <Text style={styles.challengeText}>Valor: {challenge.stakeAmount} Palpicoins</Text>
                  <Text style={styles.challengeText}>Status: {challengeStatusLabel(challenge.status)}</Text>
                  {challenge.status === 'SETTLED' &&
                  challenge.creatorPoints != null &&
                  challenge.opponentPoints != null ? (
                    <Text style={styles.challengeResult}>
                      {scoreLabel(challenge, user?.id, profileQuery.data.name)}
                    </Text>
                  ) : null}
                </View>
              ))}
            </View>
          </>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.statCard}>
      <Text style={styles.statValue}>{value}</Text>
      <Text style={styles.statLabel}>{label}</Text>
    </View>
  );
}

function formatDate(value: string | null) {
  if (!value) {
    return '-';
  }
  return new Intl.DateTimeFormat('pt-BR', { month: 'short', year: 'numeric' }).format(new Date(value));
}

function challengeStatusLabel(status: string) {
  const labels: Record<string, string> = {
    ACCEPTED: 'Aceito',
    CANCELLED: 'Cancelado',
    DECLINED: 'Recusado',
    PENDING: 'Pendente',
    SETTLED: 'Finalizado',
  };
  return labels[status] ?? status;
}

function scoreLabel(challenge: Challenge, currentUserID?: string, friendName = 'amigo') {
  const myPoints = challenge.creatorUserId === currentUserID ? challenge.creatorPoints : challenge.opponentPoints;
  const friendPoints = challenge.creatorUserId === currentUserID ? challenge.opponentPoints : challenge.creatorPoints;
  return `Final: você ${myPoints} x ${friendPoints} ${friendName}`;
}

const styles = StyleSheet.create({
  safeArea: {
    backgroundColor: colors.background,
    flex: 1,
  },
  container: {
    backgroundColor: colors.background,
    flexGrow: 1,
    gap: spacing.xl,
    paddingHorizontal: spacing.xxl,
    paddingVertical: spacing.page,
  },
  headerRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.lg,
  },
  title: {
    color: colors.primaryText,
    fontSize: 28,
    fontWeight: '800',
  },
  profileHeader: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.lg,
    padding: spacing.xl,
  },
  profileInfo: {
    flex: 1,
    gap: spacing.xs,
  },
  name: {
    color: colors.primaryText,
    fontSize: 22,
    fontWeight: '800',
    lineHeight: 28,
  },
  memberSince: {
    color: colors.mutedText,
    fontSize: 14,
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.md,
  },
  statCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flexBasis: '47%',
    flexGrow: 1,
    gap: spacing.xs,
    minHeight: 92,
    padding: spacing.lg,
  },
  statValue: {
    color: colors.primary,
    fontSize: 24,
    fontWeight: '900',
  },
  statLabel: {
    color: colors.mutedText,
    fontSize: 13,
    fontWeight: '700',
  },
  challengePanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: spacing.md,
    padding: spacing.lg,
  },
  sectionTitle: {
    color: colors.primaryText,
    fontSize: 17,
    fontWeight: '800',
  },
  mutedText: {
    color: colors.mutedText,
    fontSize: 14,
    lineHeight: 20,
  },
  challengeRow: {
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: spacing.xs,
    padding: spacing.md,
  },
  challengeTitle: {
    color: colors.primaryText,
    fontSize: 14,
    fontWeight: '800',
  },
  challengeText: {
    color: colors.mutedText,
    fontSize: 13,
    fontWeight: '700',
  },
  challengeResult: {
    color: colors.primary,
    fontSize: 13,
    fontWeight: '900',
  },
});
