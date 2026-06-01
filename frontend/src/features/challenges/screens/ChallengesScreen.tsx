import { useState } from 'react';
import { Alert, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';
import { colors } from '../../../shared/theme';
import { PALPICOIN_NOTICE } from '../../palpicoins/services/palpicoins';
import { useAuth } from '../../auth/hooks/useAuth';
import {
  acceptChallenge,
  cancelChallenge,
  declineChallenge,
  listChallenges,
  type Challenge,
} from '../services/challenges';

type Props = {
  onBack: () => void;
};

type Tab = 'received' | 'accepted' | 'sent' | 'history';

export function ChallengesScreen({ onBack }: Props) {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const receivedQuery = useQuery({
    queryFn: () => listChallenges({ status: 'PENDING', type: 'received' }),
    queryKey: ['challenges', 'received'],
  });
  const sentQuery = useQuery({
    queryFn: () => listChallenges({ type: 'sent' }),
    queryKey: ['challenges', 'sent'],
  });
  const acceptedQuery = useQuery({
    queryFn: () => listChallenges({ type: 'all' }),
    queryKey: ['challenges', 'accepted'],
  });
  const historyQuery = useQuery({
    queryFn: () => listChallenges({ status: 'SETTLED', type: 'all' }),
    queryKey: ['challenges', 'history'],
  });
  const [activeTab, setActiveTab] = useState<Tab>('received');
  const acceptMutation = useChallengeAction(acceptChallenge, queryClient);
  const declineMutation = useChallengeAction(declineChallenge, queryClient);
  const cancelMutation = useChallengeAction(cancelChallenge, queryClient);

  const activeItems =
    activeTab === 'received'
      ? receivedQuery.data?.challenges
      : activeTab === 'accepted'
        ? acceptedQuery.data?.challenges.filter(
            (challenge) => challenge.status === 'ACCEPTED' || challenge.status === 'SETTLED',
          )
      : activeTab === 'sent'
        ? sentQuery.data?.challenges
        : historyQuery.data?.challenges;
  const isLoading =
    activeTab === 'received'
      ? receivedQuery.isLoading
      : activeTab === 'accepted'
        ? acceptedQuery.isLoading
      : activeTab === 'sent'
        ? sentQuery.isLoading
        : historyQuery.isLoading;

  return (
    <SafeAreaView style={styles.safeArea}>
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.topBar}>
          <BackButton onPress={onBack} />
          <Text style={styles.notice}>{PALPICOIN_NOTICE}</Text>
        </View>
        <Text style={styles.title}>Desafios</Text>

        <View style={styles.tabs}>
          {(['received', 'accepted', 'sent', 'history'] as const).map((tab) => (
            <Pressable
              key={tab}
              onPress={() => setActiveTab(tab)}
              style={[styles.tab, activeTab === tab && styles.tabActive]}>
              <Text style={[styles.tabText, activeTab === tab && styles.tabTextActive]}>
                {tabLabel(tab)}
              </Text>
            </Pressable>
          ))}
        </View>

        {isLoading ? <LoadingIndicator text="Carregando desafios..." /> : null}
        {!isLoading && !activeItems?.length ? (
          <EmptyBox title="Nenhum desafio" text="Seus desafios aparecerão aqui." />
        ) : null}

        {activeItems?.map((challenge) => (
          <ChallengeCard
            activeTab={activeTab}
            challenge={challenge}
            isAccepting={acceptMutation.isPending}
            isCancelling={cancelMutation.isPending}
            isDeclining={declineMutation.isPending}
            key={challenge.id}
            onAccept={() => acceptMutation.mutate(challenge.id)}
            onCancel={() => cancelMutation.mutate(challenge.id)}
            onDecline={() => declineMutation.mutate(challenge.id)}
            userID={user?.id}
          />
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

function useChallengeAction(
  action: (challengeID: string) => Promise<Challenge>,
  queryClient: ReturnType<typeof useQueryClient>,
) {
  return useMutation({
    mutationFn: action,
    onError: (error) => {
      Alert.alert('Erro', error instanceof Error ? error.message : 'Não foi possível atualizar.');
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['challenges'] });
      await queryClient.invalidateQueries({ queryKey: ['me', 'wallet'] });
      Alert.alert('Sucesso', 'Desafio atualizado.');
    },
  });
}

function ChallengeCard({
  activeTab,
  challenge,
  isAccepting,
  isCancelling,
  isDeclining,
  onAccept,
  onCancel,
  onDecline,
  userID,
}: {
  activeTab: Tab;
  challenge: Challenge;
  isAccepting: boolean;
  isCancelling: boolean;
  isDeclining: boolean;
  onAccept: () => void;
  onCancel: () => void;
  onDecline: () => void;
  userID?: string;
}) {
  const matchName =
    challenge.homeTeam && challenge.awayTeam
      ? `${challenge.homeTeam} x ${challenge.awayTeam}`
      : 'Jogo selecionado';

  return (
    <View style={styles.card}>
      <Text style={styles.cardTitle}>{challenge.friendName}</Text>
      <Text style={styles.cardText}>{matchName}</Text>
      {challenge.kickoffAt ? (
        <Text style={styles.cardText}>Jogo: {new Date(challenge.kickoffAt).toLocaleString()}</Text>
      ) : null}
      <Text style={styles.cardText}>Valor: {challenge.stakeAmount} Palpicoins</Text>
      {challenge.status === 'SETTLED' ? (
        <Text style={styles.scoreText}>{scoreLabel(challenge, userID)}</Text>
      ) : null}
      {activeTab === 'history' ? (
        <Text style={styles.resultText}>{historyLabel(challenge, userID)}</Text>
      ) : null}
      {activeTab === 'accepted' && challenge.status === 'ACCEPTED' ? (
        <Text style={styles.statusText}>Aceito, aguardando encerramento do jogo</Text>
      ) : null}
      {activeTab === 'sent' ? <Text style={styles.statusText}>{statusLabel(challenge.status)}</Text> : null}
      {activeTab === 'received' ? (
        <View style={styles.actions}>
          <Pressable
            disabled={isAccepting}
            onPress={onAccept}
            style={[styles.actionButton, styles.acceptButton]}>
            <Text style={styles.acceptButtonText}>Aceitar</Text>
          </Pressable>
          <Pressable
            disabled={isDeclining}
            onPress={onDecline}
            style={[styles.actionButton, styles.declineButton]}>
            <Text style={styles.declineButtonText}>Recusar</Text>
          </Pressable>
        </View>
      ) : null}
      {activeTab === 'sent' && challenge.status === 'PENDING' ? (
        <Pressable disabled={isCancelling} onPress={onCancel} style={styles.cancelButton}>
          <Text style={styles.declineButtonText}>Cancelar</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

function tabLabel(tab: Tab) {
  if (tab === 'received') {
    return 'Recebidos';
  }
  if (tab === 'accepted') {
    return 'Aceitos';
  }
  if (tab === 'sent') {
    return 'Enviados';
  }
  return 'Histórico';
}

function statusLabel(status: Challenge['status']) {
  const labels: Record<Challenge['status'], string> = {
    ACCEPTED: 'Aceito',
    CANCELLED: 'Cancelado',
    DECLINED: 'Recusado',
    PENDING: 'Pendente',
    SETTLED: 'Finalizado',
  };
  return labels[status];
}

function historyLabel(challenge: Challenge, userID?: string) {
  const opponent = challenge.friendName ? ` contra ${challenge.friendName}` : '';
  if (!challenge.winnerUserId) {
    return `Empatou${opponent}: ${challenge.stakeAmount} Palpicoins devolvidos`;
  }
  const won = challenge.winnerUserId === userID;
  return won
    ? `Venceu${opponent}: +${challenge.stakeAmount} Palpicoins`
    : `Perdeu${opponent}: -${challenge.stakeAmount} Palpicoins`;
}

function scoreLabel(challenge: Challenge, userID?: string) {
  if (challenge.creatorPoints == null || challenge.opponentPoints == null) {
    return 'Pontuação final indisponível';
  }
  const myPoints = challenge.creatorUserId === userID ? challenge.creatorPoints : challenge.opponentPoints;
  const friendPoints = challenge.creatorUserId === userID ? challenge.opponentPoints : challenge.creatorPoints;
  return `Pontuação final: você ${myPoints} x ${friendPoints} ${challenge.friendName}`;
}

const styles = StyleSheet.create({
  safeArea: { backgroundColor: colors.background, flex: 1 },
  container: { gap: 16, paddingHorizontal: 24, paddingVertical: 32 },
  topBar: { alignItems: 'center', flexDirection: 'row', gap: 12 },
  notice: { color: colors.mutedText, flex: 1, fontSize: 12, lineHeight: 17 },
  title: { color: colors.primaryDark, fontSize: 32, fontWeight: '900' },
  tabs: { backgroundColor: '#edf3e8', borderRadius: 8, flexDirection: 'row', gap: 6, padding: 6 },
  tab: { alignItems: 'center', borderRadius: 7, flex: 1, minHeight: 42, justifyContent: 'center' },
  tabActive: { backgroundColor: colors.surface },
  tabText: { color: colors.mutedText, fontSize: 12, fontWeight: '900' },
  tabTextActive: { color: colors.primary },
  card: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 16,
  },
  cardTitle: { color: colors.primaryDark, fontSize: 16, fontWeight: '900' },
  cardText: { color: colors.mutedText, fontSize: 13, fontWeight: '700' },
  resultText: { color: colors.primary, fontSize: 14, fontWeight: '900' },
  scoreText: { color: colors.primaryDark, fontSize: 13, fontWeight: '900' },
  statusText: { color: colors.primary, fontSize: 13, fontWeight: '900' },
  actions: { flexDirection: 'row', gap: 10, marginTop: 8 },
  actionButton: { alignItems: 'center', borderRadius: 8, flex: 1, minHeight: 44, justifyContent: 'center' },
  acceptButton: { backgroundColor: colors.primary },
  acceptButtonText: { color: colors.white, fontSize: 14, fontWeight: '900' },
  declineButton: { backgroundColor: colors.surface, borderColor: colors.danger, borderWidth: 1 },
  declineButtonText: { color: colors.danger, fontSize: 14, fontWeight: '900' },
  cancelButton: {
    alignItems: 'center',
    borderColor: colors.danger,
    borderRadius: 8,
    borderWidth: 1,
    justifyContent: 'center',
    marginTop: 8,
    minHeight: 44,
  },
});
