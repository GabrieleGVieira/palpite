import { useCallback, useEffect, useState } from 'react';
import { Image, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { colors } from '../../../shared/theme';
import { listGroupMembers, type Group, type GroupMember } from '../services/groups';

type Props = {
  group: Group;
  onBack: () => void;
  onOpenMember: (member: GroupMember) => void;
};

export function GroupMembersScreen({ group, onBack, onOpenMember }: Props) {
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadMembers = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      setMembers(await listGroupMembers(group.id));
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : 'Não foi possível carregar os participantes.',
      );
    } finally {
      setIsLoading(false);
    }
  }, [group.id]);

  useEffect(() => {
    void loadMembers();
  }, [loadMembers]);

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.topBar}>
          <BackButton onPress={onBack} />
        </View>

        <View>
          <Text style={styles.title}>Participantes</Text>
          <Text style={styles.subtitle}>{group.name}</Text>
        </View>

        {isLoading ? <LoadingIndicator text="Carregando participantes..." /> : null}

        {!isLoading && error ? (
          <View style={styles.errorBox}>
            <Text style={styles.errorText}>{error}</Text>
            <Pressable onPress={loadMembers} style={styles.retryButton}>
              <Text style={styles.retryText}>Tentar novamente</Text>
            </Pressable>
          </View>
        ) : null}

        {!isLoading && !error && members.length === 0 ? (
          <EmptyBox title="Nenhum participante" text="Este grupo ainda não tem membros ativos." />
        ) : null}

        {!isLoading && !error
          ? members.map((member) => (
              <Pressable
                key={member.user_id}
                onPress={() => onOpenMember(member)}
                style={styles.memberCard}>
                <Avatar avatarURL={member.avatar_url} name={member.display_name} />
                <View style={styles.memberContent}>
                  <View style={styles.memberHeader}>
                    <Text numberOfLines={1} style={styles.memberName}>
                      {displayName(member)}
                    </Text>
                    <RoleBadge role={member.role} />
                  </View>
                  <Text style={styles.memberMeta}>
                    {member.ranking ? `#${member.ranking}` : 'Sem ranking'} • {member.points ?? 0}{' '}
                    pts
                  </Text>
                  <Text style={styles.joinedText}>Entrou em {formatDate(member.joined_at)}</Text>
                </View>
              </Pressable>
            ))
          : null}
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

function RoleBadge({ role }: { role: string }) {
  return (
    <View style={[styles.roleBadge, role === 'owner' && styles.ownerBadge]}>
      <Text style={[styles.roleText, role === 'owner' && styles.ownerText]}>{roleLabel(role)}</Text>
    </View>
  );
}

function displayName(member: GroupMember) {
  return member.display_name || `Usuário ${member.user_id.slice(0, 8)}`;
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
    gap: 16,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  topBar: {
    alignItems: 'flex-start',
  },
  title: {
    color: colors.primaryText,
    fontSize: 30,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.mutedText,
    fontSize: 15,
    lineHeight: 22,
    marginTop: 6,
  },
  errorBox: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: 12,
    padding: 16,
  },
  errorText: {
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
  },
  retryButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    minHeight: 42,
    justifyContent: 'center',
  },
  retryText: {
    color: colors.surface,
    fontSize: 13,
    fontWeight: '800',
  },
  memberCard: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    gap: 12,
    padding: 14,
  },
  avatar: {
    borderRadius: 24,
    height: 48,
    width: 48,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    borderRadius: 24,
    height: 48,
    justifyContent: 'center',
    width: 48,
  },
  avatarText: {
    color: colors.primary,
    fontSize: 15,
    fontWeight: '800',
  },
  memberContent: {
    flex: 1,
    gap: 4,
  },
  memberHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 8,
  },
  memberName: {
    color: colors.primaryText,
    flex: 1,
    fontSize: 16,
    fontWeight: '800',
  },
  roleBadge: {
    backgroundColor: '#eef3ea',
    borderRadius: 999,
    paddingHorizontal: 8,
    paddingVertical: 4,
  },
  ownerBadge: {
    backgroundColor: '#dff0e3',
  },
  roleText: {
    color: colors.mutedText,
    fontSize: 11,
    fontWeight: '800',
  },
  ownerText: {
    color: colors.primary,
  },
  memberMeta: {
    color: colors.primary,
    fontSize: 13,
    fontWeight: '800',
  },
  joinedText: {
    color: colors.mutedText,
    fontSize: 12,
  },
});
