import { StatusBar } from 'expo-status-bar';
import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Image,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth } from '../hooks/useAuth';
import { joinGroup, listGroups, type Group } from '../services/groups';

type HomeScreenProps = {
  onCreateGroup: () => void;
};

export function HomeScreen({ onCreateGroup }: HomeScreenProps) {
  const { isSubmitting, logout, user } = useAuth();
  const userName = user?.user_metadata.full_name as string | undefined;
  const [groups, setGroups] = useState<Group[]>([]);
  const [isLoadingGroups, setIsLoadingGroups] = useState(true);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [inviteCode, setInviteCode] = useState('');
  const [joinError, setJoinError] = useState<string | null>(null);
  const [isJoiningGroup, setIsJoiningGroup] = useState(false);

  const loadGroups = useCallback(async () => {
    setGroupsError(null);
    setIsLoadingGroups(true);

    try {
      const nextGroups = await listGroups();
      setGroups(nextGroups);
    } catch (error) {
      setGroupsError(
        error instanceof Error ? error.message : 'Nao foi possivel carregar seus grupos.',
      );
    } finally {
      setIsLoadingGroups(false);
    }
  }, []);

  useEffect(() => {
    loadGroups();
  }, [loadGroups]);

  async function handleJoinGroup() {
    setJoinError(null);

    if (!inviteCode.trim()) {
      setJoinError('Informe o codigo do grupo.');
      return;
    }

    setIsJoiningGroup(true);

    try {
      await joinGroup(inviteCode);
      setInviteCode('');
      await loadGroups();
    } catch (error) {
      setJoinError(error instanceof Error ? error.message : 'Nao foi possivel entrar no grupo.');
    } finally {
      setIsJoiningGroup(false);
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <View style={styles.header}>
          <View style={styles.logoMark}>
            <Image
              accessibilityIgnoresInvertColors
              resizeMode="cover"
              source={require('../../assets/splash-palpitai.png')}
              style={styles.logoImage}
            />
          </View>
          <Text style={styles.title}>PalpitAI</Text>
          <Text style={styles.subtitle}>
            Sessao ativa via Supabase Auth. Seus grupos, palpites e ranking entram aqui.
          </Text>
        </View>

        <View style={styles.sessionBox}>
          <Text style={styles.sessionLabel}>Conta conectada</Text>
          <Text style={styles.sessionName}>{userName || user?.email || 'Usuario'}</Text>
          {user?.email ? <Text style={styles.sessionEmail}>{user.email}</Text> : null}
        </View>

        <View style={styles.joinBox}>
          <View>
            <Text style={styles.joinTitle}>Entrar em um grupo</Text>
            <Text style={styles.joinSubtitle}>Use o codigo de convite recebido.</Text>
          </View>

          <View style={styles.joinForm}>
            <TextInput
              autoCapitalize="characters"
              onChangeText={setInviteCode}
              placeholder="CODIGO"
              placeholderTextColor="#7c8898"
              style={styles.inviteInput}
              value={inviteCode}
            />
            <Pressable
              disabled={isJoiningGroup}
              onPress={handleJoinGroup}
              style={[styles.joinButton, isJoiningGroup && styles.buttonDisabled]}>
              <Text style={styles.joinButtonText}>{isJoiningGroup ? 'Entrando...' : 'Entrar'}</Text>
            </Pressable>
          </View>

          {joinError ? <Text style={styles.errorText}>{joinError}</Text> : null}
        </View>

        <View style={styles.groupsSection}>
          <View style={styles.sectionHeader}>
            <View>
              <Text style={styles.sectionTitle}>Meus grupos</Text>
              <Text style={styles.sectionSubtitle}>Bolões em que voce participa</Text>
            </View>
            <Pressable onPress={loadGroups} style={styles.refreshButton}>
              <Text style={styles.refreshButtonText}>Atualizar</Text>
            </Pressable>
          </View>

          {isLoadingGroups ? (
            <View style={styles.loadingBox}>
              <ActivityIndicator color="#1f7a4a" />
              <Text style={styles.loadingText}>Carregando grupos...</Text>
            </View>
          ) : null}

          {groupsError ? <Text style={styles.errorText}>{groupsError}</Text> : null}

          {!isLoadingGroups && !groupsError && groups.length === 0 ? (
            <View style={styles.emptyBox}>
              <Text style={styles.emptyTitle}>Nenhum grupo ainda</Text>
              <Text style={styles.emptyText}>
                Crie seu primeiro bolao da Copa para convidar sua turma.
              </Text>
            </View>
          ) : null}

          {groups.map((group) => (
            <View key={group.id} style={styles.groupCard}>
              <View style={styles.groupCardHeader}>
                <View style={styles.groupTitleBlock}>
                  <Text style={styles.groupName}>{group.name}</Text>
                  <Text style={styles.groupMeta}>
                    {group.role === 'owner' ? 'Dono' : 'Membro'} · {group.member_count} participante
                    {group.member_count === 1 ? '' : 's'}
                  </Text>
                </View>
                <View style={styles.inviteBadge}>
                  <Text style={styles.inviteBadgeLabel}>Convite</Text>
                  <Text style={styles.inviteBadgeCode}>{group.invite_code}</Text>
                </View>
              </View>

              {group.description ? (
                <Text style={styles.groupDescription}>{group.description}</Text>
              ) : null}

              <Text style={styles.groupScope}>
                {group.match_scope === 'all'
                  ? 'Todos os jogos da Copa'
                  : `Selecoes: ${group.selected_teams.join(', ')}`}
              </Text>
            </View>
          ))}
        </View>

        <View style={styles.actions}>
          <Pressable onPress={onCreateGroup} style={styles.primaryButton}>
            <Text style={styles.primaryButtonText}>Criar grupo</Text>
          </Pressable>

          <Pressable disabled={isSubmitting} onPress={logout} style={styles.secondaryButton}>
            <Text style={styles.secondaryButtonText}>{isSubmitting ? 'Saindo...' : 'Sair'}</Text>
          </Pressable>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  container: {
    backgroundColor: '#f5f8ef',
    flexGrow: 1,
    gap: 24,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  pitchMarkTop: {
    borderColor: 'rgba(255, 255, 255, 0.68)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  pitchCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 140,
    borderWidth: 2,
    height: 280,
    position: 'absolute',
    right: -128,
    top: 104,
    width: 280,
  },
  header: {
    paddingTop: 32,
  },
  logoMark: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 32,
    borderWidth: 1,
    height: 64,
    justifyContent: 'center',
    marginBottom: 24,
    overflow: 'hidden',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 8, width: 0 },
    shadowOpacity: 0.12,
    shadowRadius: 16,
    width: 64,
  },
  logoImage: {
    height: 76,
    transform: [{ scale: 1.18 }],
    width: 76,
  },
  title: {
    color: '#123d2a',
    fontSize: 38,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 340,
  },
  sessionBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 20,
  },
  sessionLabel: {
    color: '#486654',
    fontSize: 13,
    fontWeight: '700',
    textTransform: 'uppercase',
  },
  sessionName: {
    color: '#123d2a',
    fontSize: 24,
    fontWeight: '800',
    marginTop: 8,
  },
  sessionEmail: {
    color: '#486654',
    fontSize: 15,
    marginTop: 4,
  },
  joinBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 14,
    padding: 16,
  },
  joinTitle: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  joinSubtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  joinForm: {
    flexDirection: 'row',
    gap: 10,
  },
  inviteInput: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    flex: 1,
    fontSize: 16,
    fontWeight: '800',
    minHeight: 48,
    paddingHorizontal: 14,
  },
  joinButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 48,
    paddingHorizontal: 18,
  },
  buttonDisabled: {
    opacity: 0.72,
  },
  joinButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '800',
  },
  groupsSection: {
    gap: 12,
  },
  sectionHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  sectionTitle: {
    color: '#123d2a',
    fontSize: 22,
    fontWeight: '800',
  },
  sectionSubtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 3,
  },
  refreshButton: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  refreshButtonText: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '800',
  },
  loadingBox: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 18,
  },
  loadingText: {
    color: '#486654',
    fontSize: 14,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  emptyBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 18,
  },
  emptyTitle: {
    color: '#123d2a',
    fontSize: 17,
    fontWeight: '800',
  },
  emptyText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 6,
  },
  groupCard: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 16,
  },
  groupCardHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: 12,
    justifyContent: 'space-between',
  },
  groupTitleBlock: {
    flex: 1,
  },
  groupName: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  groupMeta: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  inviteBadge: {
    alignItems: 'center',
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  inviteBadgeLabel: {
    color: '#486654',
    fontSize: 10,
    fontWeight: '800',
    textTransform: 'uppercase',
  },
  inviteBadgeCode: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
    marginTop: 2,
  },
  groupDescription: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 12,
  },
  groupScope: {
    color: '#1f7a4a',
    fontSize: 13,
    fontWeight: '800',
    marginTop: 12,
  },
  actions: {
    gap: 12,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 54,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '800',
  },
  secondaryButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderRadius: 8,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 54,
  },
  secondaryButtonText: {
    color: '#1f7a4a',
    fontSize: 16,
    fontWeight: '800',
  },
});
