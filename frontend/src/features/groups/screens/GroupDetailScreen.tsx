import { StatusBar } from 'expo-status-bar';
import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { GroupDetailHeader } from '../components/group-details/GroupDetailHeader';
import { GroupDetailMatchCard } from '../components/group-details/GroupDetailMatchCard';
import { useGroupDetailScreen } from '../hooks/useGroupDetailScreen';
import type { Group } from '../services/groups';
import { GroupDetailRankingCard } from '../components/group-details/GroupDetailRankingCard';
import { EmptyBox } from '../../../shared/components/EmptyBox';
import { LoadingIndicator } from '../../../shared/components/LoadingIndicator';

type GroupDetailScreenProps = {
  group: Group;
  onBack: () => void;
  onGroupLeft: () => void;
  onOpenAdmin: () => void;
  onOpenMembers: () => void;
};

export function GroupDetailScreen({
  group,
  onBack,
  onGroupLeft,
  onOpenAdmin,
  onOpenMembers,
}: GroupDetailScreenProps) {
  const {
    activeTab,
    drafts,
    isLoading,
    isLoadingRanking,
    isLeavingGroup,
    leaveGroup,
    matches,
    notificationMessage,
    ranking,
    rankingError,
    savePrediction,
    setActiveTab,
    savingMatchID,
    updateDraft,
  } = useGroupDetailScreen(group, onGroupLeft);

  function confirmLeaveGroup() {
    Alert.alert(
      'Sair do grupo',
      'Você tem certeza que deseja sair deste grupo? Você não verá mais os jogos e ranking dele.',
      [
        { style: 'cancel', text: 'Cancelar' },
        { onPress: leaveGroup, style: 'destructive', text: 'Sair' },
      ],
    );
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <GroupDetailHeader
          activeTab={activeTab}
          group={group}
          isLeavingGroup={isLeavingGroup}
          notificationMessage={notificationMessage}
          onBack={onBack}
          onChangeTab={setActiveTab}
          onLeaveGroup={confirmLeaveGroup}
          onOpenAdmin={onOpenAdmin}
          onOpenMembers={onOpenMembers}
        />

        {activeTab === 'matches' && isLoading ? (
          <LoadingIndicator text="Carregando jogos..." />
        ) : null}

        {activeTab === 'matches' && !isLoading && matches.length === 0 ? (
          <EmptyBox
            title="Nenhum jogo encontrado"
            text="A lista de jogos desse bolão ainda está vazia."
          />
        ) : null}

        {activeTab === 'matches'
          ? matches.map((match) => {
              const draft = drafts[match.id] ?? { awayScore: '', homeScore: '' };
              const isSaving = savingMatchID === match.id;

              return (
                <GroupDetailMatchCard
                  key={match.id}
                  draft={draft}
                  isSaving={isSaving}
                  match={match}
                  onChangeDraft={updateDraft}
                  onSavePrediction={savePrediction}
                />
              );
            })
          : null}

        {activeTab === 'ranking' && isLoadingRanking ? (
          <LoadingIndicator text="Carregando ranking..." />
        ) : null}

        {activeTab === 'ranking' && rankingError ? (
          <Text style={styles.errorText}>{rankingError}</Text>
        ) : null}

        {activeTab === 'ranking' && !isLoadingRanking && !rankingError && ranking.length === 0 ? (
          <EmptyBox
            title="Ranking vazio"
            text="Os participantes ainda não pontuaram neste grupo."
          />
        ) : null}

        {activeTab === 'ranking'
          ? ranking.map((entry) => <GroupDetailRankingCard entry={entry} key={entry.user_id} />)
          : null}
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
    gap: 18,
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
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
});
