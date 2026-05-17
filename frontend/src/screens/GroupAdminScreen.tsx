import { StatusBar } from 'expo-status-bar';
import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Switch,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import {
  approveJoinRequest,
  listJoinRequests,
  updateGroup,
  type Group,
  type JoinRequest,
} from '../services/groups';

type GroupAdminScreenProps = {
  group: Group;
  onBack: () => void;
  onGroupUpdated: (group: Group) => void;
};

export function GroupAdminScreen({ group, onBack, onGroupUpdated }: GroupAdminScreenProps) {
  const [name, setName] = useState(group.name);
  const [description, setDescription] = useState(group.description);
  const [isPrivate, setIsPrivate] = useState(group.is_private);
  const [hasUnlimitedParticipants, setHasUnlimitedParticipants] = useState(
    group.participant_limit === null,
  );
  const [participantLimit, setParticipantLimit] = useState(
    group.participant_limit ? String(group.participant_limit) : '20',
  );
  const [requests, setRequests] = useState<JoinRequest[]>([]);
  const [isLoadingRequests, setIsLoadingRequests] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [approvingUserID, setApprovingUserID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const loadRequests = useCallback(async () => {
    setError(null);
    setIsLoadingRequests(true);

    try {
      const nextRequests = await listJoinRequests(group.id);
      setRequests(nextRequests);
    } catch (loadError) {
      setError(
        loadError instanceof Error ? loadError.message : 'Nao foi possivel carregar solicitacoes.',
      );
    } finally {
      setIsLoadingRequests(false);
    }
  }, [group.id]);

  useEffect(() => {
    loadRequests();
  }, [loadRequests]);

  async function handleSaveGroup() {
    setError(null);
    setSuccessMessage(null);

    if (!name.trim()) {
      setError('Informe o nome do grupo.');
      return;
    }

    if (!hasUnlimitedParticipants && Number(participantLimit) < 2) {
      setError('O limite precisa ser maior que 1.');
      return;
    }

    setIsSaving(true);

    try {
      const updatedGroup = await updateGroup(group.id, {
        description,
        has_unlimited_participants: hasUnlimitedParticipants,
        is_private: isPrivate,
        name,
        participant_limit: hasUnlimitedParticipants ? null : Number(participantLimit),
      });
      onGroupUpdated({ ...group, ...updatedGroup });
      setSuccessMessage('Grupo atualizado.');
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : 'Nao foi possivel atualizar o grupo.',
      );
    } finally {
      setIsSaving(false);
    }
  }

  async function handleApprove(request: JoinRequest) {
    setError(null);
    setSuccessMessage(null);
    setApprovingUserID(request.user_id);

    try {
      await approveJoinRequest(group.id, request.user_id);
      setRequests((currentRequests) =>
        currentRequests.filter((currentRequest) => currentRequest.user_id !== request.user_id),
      );
      onGroupUpdated({
        ...group,
        member_count: group.member_count + 1,
        pending_requests_count: Math.max(group.pending_requests_count - 1, 0),
      });
      setSuccessMessage('Solicitacao aprovada.');
    } catch (approveError) {
      setError(
        approveError instanceof Error
          ? approveError.message
          : 'Nao foi possivel aprovar a solicitacao.',
      );
    } finally {
      setApprovingUserID(null);
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <Pressable accessibilityLabel="Voltar" onPress={onBack} style={styles.backButton}>
          <Text style={styles.backButtonText}>‹</Text>
        </Pressable>

        <View>
          <Text style={styles.title}>Admin do grupo</Text>
          <Text style={styles.subtitle}>{group.name}</Text>
        </View>

        {error ? <Text style={styles.errorText}>{error}</Text> : null}
        {successMessage ? <Text style={styles.successText}>{successMessage}</Text> : null}

        <View style={styles.card}>
          <Text style={styles.cardTitle}>Informacoes</Text>

          <View style={styles.fieldGroup}>
            <Text style={styles.label}>Nome</Text>
            <TextInput
              autoCapitalize="words"
              onChangeText={setName}
              placeholder="Nome do grupo"
              placeholderTextColor="#7c8898"
              style={styles.input}
              value={name}
            />
          </View>

          <View style={styles.fieldGroup}>
            <Text style={styles.label}>Descricao</Text>
            <TextInput
              multiline
              onChangeText={setDescription}
              placeholder="Descricao do grupo"
              placeholderTextColor="#7c8898"
              style={[styles.input, styles.textArea]}
              textAlignVertical="top"
              value={description}
            />
          </View>

          <View style={styles.row}>
            <View style={[styles.fieldGroup, styles.limitField]}>
              <Text style={styles.label}>Participantes</Text>
              <TextInput
                editable={!hasUnlimitedParticipants}
                keyboardType="number-pad"
                onChangeText={setParticipantLimit}
                placeholder={hasUnlimitedParticipants ? 'Ilimitado' : '20'}
                placeholderTextColor="#7c8898"
                style={[styles.input, hasUnlimitedParticipants && styles.inputDisabled]}
                value={hasUnlimitedParticipants ? '' : participantLimit}
              />
            </View>

            <View style={styles.switchBox}>
              <View>
                <Text style={styles.switchTitle}>Ilimitado</Text>
                <Text style={styles.switchSubtitle}>Sem teto</Text>
              </View>
              <Switch
                onValueChange={setHasUnlimitedParticipants}
                thumbColor="#ffffff"
                trackColor={{ false: '#c9d7c3', true: '#1f7a4a' }}
                value={hasUnlimitedParticipants}
              />
            </View>
          </View>

          <View style={styles.switchBoxWide}>
            <View>
              <Text style={styles.switchTitle}>Privado</Text>
              <Text style={styles.switchSubtitle}>Novos membros precisam de aprovacao</Text>
            </View>
            <Switch
              onValueChange={setIsPrivate}
              thumbColor="#ffffff"
              trackColor={{ false: '#c9d7c3', true: '#1f7a4a' }}
              value={isPrivate}
            />
          </View>

          <Pressable
            disabled={isSaving}
            onPress={handleSaveGroup}
            style={[styles.primaryButton, isSaving && styles.buttonDisabled]}>
            <Text style={styles.primaryButtonText}>{isSaving ? 'Salvando...' : 'Salvar'}</Text>
          </Pressable>
        </View>

        <View style={styles.card}>
          <View style={styles.requestsHeader}>
            <View>
              <Text style={styles.cardTitle}>Solicitacoes</Text>
              <Text style={styles.cardSubtitle}>Usuarios aguardando aceite</Text>
            </View>
            <Pressable onPress={loadRequests} style={styles.refreshButton}>
              <Text style={styles.refreshButtonText}>Atualizar</Text>
            </Pressable>
          </View>

          {isLoadingRequests ? (
            <View style={styles.loadingBox}>
              <ActivityIndicator color="#1f7a4a" />
              <Text style={styles.loadingText}>Carregando...</Text>
            </View>
          ) : null}

          {!isLoadingRequests && requests.length === 0 ? (
            <Text style={styles.emptyText}>Nenhuma solicitacao pendente.</Text>
          ) : null}

          {requests.map((request) => {
            const isApproving = approvingUserID === request.user_id;

            return (
              <View key={request.user_id} style={styles.requestRow}>
                <View style={styles.requestInfo}>
                  <Text style={styles.requestUser}>Usuario {request.user_id.slice(0, 8)}</Text>
                  <Text style={styles.requestMeta}>Solicitou entrada</Text>
                </View>
                <Pressable
                  disabled={isApproving}
                  onPress={() => handleApprove(request)}
                  style={[styles.approveButton, isApproving && styles.buttonDisabled]}>
                  <Text style={styles.approveButtonText}>
                    {isApproving ? 'Aprovando...' : 'Aprovar'}
                  </Text>
                </Pressable>
              </View>
            );
          })}
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
    gap: 20,
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
  title: {
    color: '#123d2a',
    fontSize: 34,
    fontWeight: '800',
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    marginTop: 8,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  successText: {
    color: '#1f7a4a',
    fontSize: 13,
    lineHeight: 18,
  },
  card: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 16,
    padding: 16,
  },
  cardTitle: {
    color: '#123d2a',
    fontSize: 18,
    fontWeight: '800',
  },
  cardSubtitle: {
    color: '#486654',
    fontSize: 13,
    marginTop: 4,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
  },
  input: {
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 16,
    minHeight: 52,
    paddingHorizontal: 14,
  },
  inputDisabled: {
    backgroundColor: '#edf3e8',
    color: '#7c8898',
  },
  textArea: {
    minHeight: 96,
    paddingTop: 12,
  },
  row: {
    flexDirection: 'row',
    gap: 12,
  },
  limitField: {
    flex: 1,
  },
  switchBox: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flex: 1.1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 78,
    paddingHorizontal: 12,
  },
  switchBoxWide: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 78,
    paddingHorizontal: 12,
  },
  switchTitle: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  switchSubtitle: {
    color: '#486654',
    fontSize: 12,
    marginTop: 4,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 52,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.72,
  },
  requestsHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  refreshButton: {
    backgroundColor: '#f5f8ef',
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
    gap: 8,
    padding: 12,
  },
  loadingText: {
    color: '#486654',
    fontSize: 14,
  },
  emptyText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
  },
  requestRow: {
    alignItems: 'center',
    borderTopColor: '#edf3e8',
    borderTopWidth: 1,
    flexDirection: 'row',
    gap: 10,
    justifyContent: 'space-between',
    paddingTop: 12,
  },
  requestInfo: {
    flex: 1,
  },
  requestUser: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  requestMeta: {
    color: '#486654',
    fontSize: 12,
    marginTop: 3,
  },
  approveButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 40,
    paddingHorizontal: 12,
  },
  approveButtonText: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '800',
  },
});
