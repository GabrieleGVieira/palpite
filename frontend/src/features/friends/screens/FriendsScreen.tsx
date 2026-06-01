import { useMemo, useState } from 'react';
import {
  Alert,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { NotificationBanner } from '../../../shared/components/NotificationBanner';
import { colors, spacing } from '../../../shared/theme';
import { useTemporaryNotification } from '../../../shared/hooks/useTemporaryNotification';
import { UserAvatar } from '../components/UserAvatar';
import {
  acceptFriendRequest,
  declineFriendRequest,
  listFriendRequests,
  listFriends,
  removeFriend,
  searchUsers,
  sendFriendRequest,
} from '../services/friends';
import type { Friend, FriendRequest, UserSearchResult } from '../types';

type FriendsTab = 'friends' | 'requests' | 'add';

type Props = {
  onBack: () => void;
  onOpenProfile: (userID: string) => void;
};

export function FriendsScreen({ onBack, onOpenProfile }: Props) {
  const [activeTab, setActiveTab] = useState<FriendsTab>('friends');
  const [query, setQuery] = useState('');
  const [pendingUserID, setPendingUserID] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const notification = useTemporaryNotification();

  const friendsQuery = useQuery({ queryFn: listFriends, queryKey: ['friends'] });
  const requestsQuery = useQuery({
    queryFn: listFriendRequests,
    queryKey: ['friends', 'requests'],
  });
  const searchQuery = useQuery({
    enabled: activeTab === 'add',
    queryFn: () => searchUsers(query),
    queryKey: ['users', 'search', query],
  });

  const refreshAll = async () => {
    await Promise.all([friendsQuery.refetch(), requestsQuery.refetch()]);
    if (activeTab === 'add') {
      await searchQuery.refetch();
    }
  };

  const acceptMutation = useMutation({
    mutationFn: acceptFriendRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['friends'] });
      notification.showNotification('Solicitação aceita.');
    },
  });

  const declineMutation = useMutation({
    mutationFn: declineFriendRequest,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['friends', 'requests'] });
      notification.showNotification('Solicitação recusada.');
    },
  });

  const removeMutation = useMutation({
    mutationFn: removeFriend,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['friends'] });
      await queryClient.invalidateQueries({ queryKey: ['challenges'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'wallet'] });
      notification.showNotification('Amizade removida.');
    },
  });

  const sendMutation = useMutation({
    mutationFn: sendFriendRequest,
    onMutate: (userID) => setPendingUserID(userID),
    onSettled: () => setPendingUserID(null),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['users', 'search'] });
      notification.showNotification('Solicitação enviada.');
    },
  });

  const isRefreshing = friendsQuery.isRefetching || requestsQuery.isRefetching;
  const trimmedQuery = useMemo(() => query.trim(), [query]);

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView
        contentContainerStyle={styles.container}
        refreshControl={<RefreshControl refreshing={isRefreshing} onRefresh={refreshAll} />}
        showsVerticalScrollIndicator={false}>
        <View style={styles.headerRow}>
          <BackButton onPress={onBack} />
          <View style={styles.headerText}>
            <Text style={styles.title}>Amigos</Text>
            <Text style={styles.subtitle}>Conecte-se com outros Palpiteiros.</Text>
          </View>
        </View>

        <View style={styles.tabs}>
          <TabButton active={activeTab === 'friends'} label="Meus amigos" onPress={() => setActiveTab('friends')} />
          <TabButton active={activeTab === 'requests'} label="Solicitações" onPress={() => setActiveTab('requests')} />
          <TabButton active={activeTab === 'add'} label="Adicionar" onPress={() => setActiveTab('add')} />
        </View>

        <NotificationBanner message={notification.notificationMessage} />

        {activeTab === 'friends' ? (
          <FriendsList
            error={friendsQuery.error}
            friends={friendsQuery.data ?? []}
            isLoading={friendsQuery.isLoading}
            onOpenProfile={onOpenProfile}
            onRemove={(friend) => {
              Alert.alert('Remover amizade', `Remover ${friend.name} dos seus amigos?`, [
                { style: 'cancel', text: 'Cancelar' },
                { style: 'destructive', text: 'Remover', onPress: () => removeMutation.mutate(friend.id) },
              ]);
            }}
            removingID={removeMutation.variables}
          />
        ) : null}

        {activeTab === 'requests' ? (
          <RequestsList
            acceptingID={acceptMutation.variables}
            decliningID={declineMutation.variables}
            error={requestsQuery.error}
            isLoading={requestsQuery.isLoading}
            onAccept={(request) => acceptMutation.mutate(request.id)}
            onDecline={(request) => declineMutation.mutate(request.id)}
            requests={requestsQuery.data ?? []}
          />
        ) : null}

        {activeTab === 'add' ? (
          <View style={styles.section}>
            <TextInput
              autoCapitalize="words"
              onChangeText={setQuery}
              placeholder="Buscar por nome"
              placeholderTextColor={colors.mutedText}
              style={styles.searchInput}
              value={query}
            />
            {searchQuery.isLoading ? <LoadingIndicator text="Buscando Palpiteiros..." /> : null}
            {searchQuery.error ? <Text style={styles.errorText}>Não foi possivel buscar usuarios.</Text> : null}
            {!searchQuery.isLoading && !searchQuery.error && (searchQuery.data ?? []).length === 0 ? (
              <EmptyBox
                title="Nenhum resultado"
                text={trimmedQuery ? 'Tente outro nome.' : 'Digite um nome para encontrar Palpiteiros.'}
              />
            ) : null}
            {(searchQuery.data ?? []).map((user) => (
              <SearchUserCard
                key={user.id}
                isSending={pendingUserID === user.id}
                onAdd={() => sendMutation.mutate(user.id)}
                user={user}
              />
            ))}
          </View>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function FriendsList({
  error,
  friends,
  isLoading,
  onOpenProfile,
  onRemove,
  removingID,
}: {
  error: Error | null;
  friends: Friend[];
  isLoading: boolean;
  onOpenProfile: (userID: string) => void;
  onRemove: (friend: Friend) => void;
  removingID?: string;
}) {
  if (isLoading) {
    return <LoadingIndicator text="Carregando amigos..." />;
  }
  if (error) {
    return <Text style={styles.errorText}>Não foi possivel carregar seus amigos.</Text>;
  }
  if (friends.length === 0) {
    return <EmptyBox title="Você ainda não possui amigos." text="Busque Palpiteiros e envie solicitações." />;
  }
  return (
    <View style={styles.section}>
      {friends.map((friend) => (
        <PersonCard
          avatarURL={friend.avatarUrl}
          key={friend.id}
          name={friend.name}
          primaryLabel="Ver perfil"
          onPrimary={() => onOpenProfile(friend.userId)}
          secondaryLabel={removingID === friend.id ? 'Removendo...' : 'Remover'}
          onSecondary={() => onRemove(friend)}
        />
      ))}
    </View>
  );
}

function RequestsList({
  acceptingID,
  decliningID,
  error,
  isLoading,
  onAccept,
  onDecline,
  requests,
}: {
  acceptingID?: string;
  decliningID?: string;
  error: Error | null;
  isLoading: boolean;
  onAccept: (request: FriendRequest) => void;
  onDecline: (request: FriendRequest) => void;
  requests: FriendRequest[];
}) {
  if (isLoading) {
    return <LoadingIndicator text="Carregando solicitações..." />;
  }
  if (error) {
    return <Text style={styles.errorText}>Não foi possivel carregar as solicitacoes.</Text>;
  }
  if (requests.length === 0) {
    return <EmptyBox title="Sem solicitações" text="Novos convites aparecerão aqui." />;
  }
  return (
    <View style={styles.section}>
      {requests.map((request) => (
        <PersonCard
          avatarURL={request.avatarUrl}
          key={request.id}
          name={request.name}
          primaryLabel={acceptingID === request.id ? 'Aceitando...' : 'Aceitar'}
          onPrimary={() => onAccept(request)}
          secondaryLabel={decliningID === request.id ? 'Recusando...' : 'Recusar'}
          onSecondary={() => onDecline(request)}
        />
      ))}
    </View>
  );
}

function SearchUserCard({
  isSending,
  onAdd,
  user,
}: {
  isSending: boolean;
  onAdd: () => void;
  user: UserSearchResult;
}) {
  const isPending = user.friendshipStatus === 'PENDING';
  const isAccepted = user.friendshipStatus === 'ACCEPTED';
  const isBlocked = user.friendshipStatus === 'BLOCKED';
  const disabled = isSending || isPending || isAccepted || isBlocked;
  const label = isSending
    ? 'Enviando...'
    : isAccepted
      ? 'Amigo'
      : isPending
        ? 'Solicitação enviada'
        : isBlocked
          ? 'Indisponível'
          : 'Adicionar amigo';

  return (
    <PersonCard
      avatarURL={user.avatarUrl}
      name={user.name}
      primaryDisabled={disabled}
      primaryLabel={label}
      onPrimary={onAdd}
    />
  );
}

function PersonCard({
  avatarURL,
  name,
  onPrimary,
  onSecondary,
  primaryDisabled,
  primaryLabel,
  secondaryLabel,
}: {
  avatarURL?: string | null;
  name: string;
  onPrimary: () => void;
  onSecondary?: () => void;
  primaryDisabled?: boolean;
  primaryLabel: string;
  secondaryLabel?: string;
}) {
  return (
    <View style={styles.personCard}>
      <View style={styles.personInfo}>
        <UserAvatar name={name} uri={avatarURL} />
        <Text numberOfLines={2} style={styles.personName}>{name}</Text>
      </View>
      <View style={styles.actions}>
        <Pressable
          disabled={primaryDisabled}
          onPress={onPrimary}
          style={({ pressed }) => [
            styles.primaryButton,
            pressed && styles.pressed,
            primaryDisabled && styles.disabledButton,
          ]}>
          <Text style={styles.primaryButtonText}>{primaryLabel}</Text>
        </Pressable>
        {secondaryLabel && onSecondary ? (
          <Pressable onPress={onSecondary} style={({ pressed }) => [styles.secondaryButton, pressed && styles.pressed]}>
            <Text style={styles.secondaryButtonText}>{secondaryLabel}</Text>
          </Pressable>
        ) : null}
      </View>
    </View>
  );
}

function TabButton({ active, label, onPress }: { active: boolean; label: string; onPress: () => void }) {
  return (
    <Pressable onPress={onPress} style={[styles.tabButton, active && styles.tabButtonActive]}>
      <Text style={[styles.tabButtonText, active && styles.tabButtonTextActive]}>{label}</Text>
    </Pressable>
  );
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
  headerText: {
    flex: 1,
  },
  title: {
    color: colors.primaryText,
    fontSize: 28,
    fontWeight: '800',
  },
  subtitle: {
    color: colors.mutedText,
    fontSize: 14,
    lineHeight: 20,
    marginTop: 4,
  },
  tabs: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    padding: 4,
  },
  tabButton: {
    alignItems: 'center',
    borderRadius: 8,
    flex: 1,
    justifyContent: 'center',
    minHeight: 42,
    paddingHorizontal: 6,
  },
  tabButtonActive: {
    backgroundColor: colors.primary,
  },
  tabButtonText: {
    color: colors.mutedText,
    fontSize: 12,
    fontWeight: '800',
    textAlign: 'center',
  },
  tabButtonTextActive: {
    color: colors.white,
  },
  section: {
    gap: spacing.md,
  },
  personCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: spacing.md,
    padding: spacing.lg,
  },
  personInfo: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
  },
  personName: {
    color: colors.primaryText,
    flex: 1,
    fontSize: 16,
    fontWeight: '800',
    lineHeight: 21,
  },
  actions: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 42,
    paddingHorizontal: 14,
  },
  primaryButtonText: {
    color: colors.white,
    fontSize: 13,
    fontWeight: '800',
  },
  secondaryButton: {
    alignItems: 'center',
    backgroundColor: colors.fieldBackground,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 42,
    paddingHorizontal: 14,
  },
  secondaryButtonText: {
    color: colors.dangerStrong,
    fontSize: 13,
    fontWeight: '800',
  },
  searchInput: {
    backgroundColor: colors.surface,
    borderColor: colors.fieldBorder,
    borderRadius: 8,
    borderWidth: 1,
    color: colors.primaryText,
    fontSize: 16,
    minHeight: 52,
    paddingHorizontal: 14,
  },
  errorText: {
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
  },
  disabledButton: {
    backgroundColor: '#8a9490',
  },
  pressed: {
    opacity: 0.86,
  },
});
