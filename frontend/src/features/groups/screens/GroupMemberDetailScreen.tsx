import { useCallback, useEffect, useState } from 'react';
import { Image, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';

import { BackButton } from '../../../shared/components/BackButton';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { colors } from '../../../shared/theme';
import {
  getGroupMemberDetail,
  type Group,
  type GroupMember,
  type GroupMemberDetail,
} from '../services/groups';

type Props = {
  group: Group;
  member: GroupMember;
  onBack: () => void;
};

export function GroupMemberDetailScreen({ group, member, onBack }: Props) {
  const [detail, setDetail] = useState<GroupMemberDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadDetail = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      setDetail(await getGroupMemberDetail(group.id, member.user_id));
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : 'Não foi possível carregar o participante.',
      );
    } finally {
      setIsLoading(false);
    }
  }, [group.id, member.user_id]);

  useEffect(() => {
    void loadDetail();
  }, [loadDetail]);

  const visibleDetail = detail ?? {
    ...member,
    accuracy_percentage: null,
    correct_predictions: null,
    predictions_count: null,
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <BackButton onPress={onBack} />

        {isLoading ? <LoadingIndicator text="Carregando participante..." /> : null}

        {!isLoading && error ? <Text style={styles.errorText}>{error}</Text> : null}

        {!isLoading && !error ? (
          <>
            <View style={styles.profileHeader}>
              <Avatar avatarURL={visibleDetail.avatar_url} name={visibleDetail.display_name} />
              <Text style={styles.name}>{displayName(visibleDetail)}</Text>
              <View style={styles.roleBadge}>
                <Text style={styles.roleText}>{roleLabel(visibleDetail.role)}</Text>
              </View>
              <Text style={styles.groupText}>{group.name}</Text>
              <Text style={styles.joinedText}>Entrou em {formatDate(visibleDetail.joined_at)}</Text>
            </View>

            <View style={styles.statsGrid}>
              <StatCard label="Posição" value={formatOptional(visibleDetail.ranking, '#')} />
              <StatCard label="Pontos" value={`${visibleDetail.points ?? 0}`} />
              <StatCard label="Palpites" value={formatOptional(visibleDetail.predictions_count)} />
              <StatCard label="Acertos" value={formatOptional(visibleDetail.correct_predictions)} />
              <StatCard
                label="Aproveitamento"
                value={
                  visibleDetail.accuracy_percentage == null
                    ? '—'
                    : `${visibleDetail.accuracy_percentage.toFixed(0)}%`
                }
              />
            </View>
          </>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function Avatar({ avatarURL, name }: { avatarURL: string | null; name: string }) {
  if (avatarURL) {
    return <Image source={{ uri: avatarURL }} style={styles.avatar} />;
  }

  return (
    <View style={styles.avatarFallback}>
      <Text style={styles.avatarText}>{initials(name)}</Text>
    </View>
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

function displayName(member: Pick<GroupMember, 'display_name' | 'user_id'>) {
  return member.display_name || `Usuário ${member.user_id.slice(0, 8)}`;
}

function formatOptional(value: number | null | undefined, prefix = '') {
  if (value == null) {
    return '—';
  }

  return `${prefix}${value}`;
}

function initials(name: string) {
  const trimmed = name.trim();
  if (!trimmed) {
    return '?';
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase();
}

function roleLabel(role: string) {
  if (role === 'owner') {
    return 'Owner';
  }
  if (role === 'admin') {
    return 'Admin';
  }

  return 'Membro';
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short' }).format(new Date(value));
}

const styles = StyleSheet.create({
  safeArea: {
    backgroundColor: colors.background,
    flex: 1,
  },
  container: {
    backgroundColor: colors.background,
    flexGrow: 1,
    gap: 18,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  errorText: {
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
  },
  profileHeader: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    padding: 20,
  },
  avatar: {
    borderRadius: 44,
    height: 88,
    width: 88,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    borderRadius: 44,
    height: 88,
    justifyContent: 'center',
    width: 88,
  },
  avatarText: {
    color: colors.primary,
    fontSize: 28,
    fontWeight: '800',
  },
  name: {
    color: colors.primaryText,
    fontSize: 24,
    fontWeight: '800',
    marginTop: 14,
    textAlign: 'center',
  },
  roleBadge: {
    backgroundColor: '#dff0e3',
    borderRadius: 999,
    marginTop: 10,
    paddingHorizontal: 10,
    paddingVertical: 5,
  },
  roleText: {
    color: colors.primary,
    fontSize: 12,
    fontWeight: '800',
  },
  groupText: {
    color: colors.mutedText,
    fontSize: 14,
    marginTop: 12,
    textAlign: 'center',
  },
  joinedText: {
    color: colors.mutedText,
    fontSize: 12,
    marginTop: 4,
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
  },
  statCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    minHeight: 86,
    padding: 14,
    width: '48%',
  },
  statValue: {
    color: colors.primaryText,
    fontSize: 22,
    fontWeight: '800',
  },
  statLabel: {
    color: colors.mutedText,
    fontSize: 12,
    marginTop: 6,
  },
});
