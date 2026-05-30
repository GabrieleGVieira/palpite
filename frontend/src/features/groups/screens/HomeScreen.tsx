import { StatusBar } from 'expo-status-bar';
import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useState } from 'react';

import { AccountSettingsCard } from '../../account/components/AccountSettingsCard';
import { DeleteAccountModal } from '../../account/components/DeleteAccountModal';
import { deleteAccount } from '../../account/services/account';
import { NotificationBanner } from '../../../shared/components/NotificationBanner';
import { LegalLinksCard } from '../../../shared/components/LegalLinksCard';
import { useAuth } from '../../auth/hooks/useAuth';
import { useHomeScreen } from '../hooks/useHomeScreen';
import { GroupListSection } from '../components/group-details/GroupListSection';
import { HomeHeader } from '../components/home/HomeHeader';
import { JoinGroupCard } from '../components/home/JoinGroupCard';
import { ScoreCard } from '../components/home/ScoreCard';
import { colors } from '../../../shared/theme';

import type { Group } from '../services/groups';

type HomeScreenProps = {
  onCreateGroup: () => void;
  onOpenGroup: (group: Group) => void;
};

export function HomeScreen({ onCreateGroup, onOpenGroup }: HomeScreenProps) {
  const { isSubmitting, logout, user } = useAuth();
  const [isDeleteModalVisible, setIsDeleteModalVisible] = useState(false);
  const [isDeletingAccount, setIsDeletingAccount] = useState(false);
  const [deleteAccountError, setDeleteAccountError] = useState<string | null>(null);
  const userName = user?.user_metadata.full_name as string | undefined;
  const {
    groups,
    totalPoints,
    isLoadingGroups,
    isLoadingScore,
    groupsError,
    scoreError,
    inviteCode,
    setInviteCode,
    isJoiningGroup,
    notificationMessage,
    refreshHome,
    handleJoinGroup,
  } = useHomeScreen();

  async function handleDeleteAccount() {
    setDeleteAccountError(null);
    setIsDeletingAccount(true);

    try {
      await deleteAccount();
      setIsDeleteModalVisible(false);
      await logout();
      Alert.alert(
        'Conta excluída',
        'Sua solicitação foi recebida. A exclusão e anonimização dos dados será processada em até 30 dias.',
      );
    } catch (error) {
      setDeleteAccountError(
        error instanceof Error ? error.message : 'Não foi possível excluir sua conta agora.',
      );
    } finally {
      setIsDeletingAccount(false);
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <HomeHeader
          userName={userName}
          onCreateGroup={onCreateGroup}
          onLogout={logout}
          isSubmitting={isSubmitting}
        />

        <ScoreCard totalPoints={totalPoints} isLoading={isLoadingScore} />

        <NotificationBanner message={notificationMessage} />

        {scoreError ? <Text style={styles.messageText}>{scoreError}</Text> : null}

        <JoinGroupCard
          inviteCode={inviteCode}
          setInviteCode={setInviteCode}
          onJoinGroup={handleJoinGroup}
          isJoiningGroup={isJoiningGroup}
        />

        <GroupListSection
          groups={groups}
          isLoadingGroups={isLoadingGroups}
          groupsError={groupsError}
          onRefresh={refreshHome}
          onOpenGroup={onOpenGroup}
        />

        <LegalLinksCard />

        <AccountSettingsCard
          onDeleteAccount={() => {
            setDeleteAccountError(null);
            setIsDeleteModalVisible(true);
          }}
        />
      </ScrollView>

      <DeleteAccountModal
        error={deleteAccountError}
        isDeleting={isDeletingAccount}
        onClose={() => setIsDeleteModalVisible(false)}
        onConfirm={handleDeleteAccount}
        visible={isDeleteModalVisible}
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: colors.background,
  },
  container: {
    backgroundColor: colors.background,
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
  messageText: {
    color: colors.danger,
    fontSize: 13,
    lineHeight: 18,
    marginTop: -8,
  },
});
