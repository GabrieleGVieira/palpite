import { StatusBar } from 'expo-status-bar';
import { Alert, ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { BackButton } from '../../../shared/components/BackButton';
import { GroupAdminForm } from '../components/admin/GroupAdminForm';
import { GroupAdminHeader } from '../components/admin/GroupAdminHeader';
import { GroupAdminMembers } from '../components/admin/GroupAdminMembers';
import { GroupAdminPayments } from '../components/admin/GroupAdminPayments';
import { GroupAdminRequests } from '../components/admin/GroupAdminRequests';
import { useGroupAdminScreen } from '../hooks/useGroupAdminScreen';
import type { Group } from '../services/groups';

type GroupAdminScreenProps = {
  group: Group;
  onBack: () => void;
  onGroupUpdated: (group: Group) => void;
};

export function GroupAdminScreen({ group, onBack, onGroupUpdated }: GroupAdminScreenProps) {
  const {
    approvingUserID,
    blockPendingPredictions,
    description,
    error,
    hasUnlimitedParticipants,
    isLoadingMembers,
    isLoadingPayments,
    isLoadingRequests,
    isPaid,
    isPrivate,
    isSaving,
    loadMembers,
    loadPayments,
    loadRequests,
    members,
    name,
    participantLimit,
    paymentAmount,
    payments,
    paymentsSummary,
    removingUserID,
    requests,
    setBlockPendingPredictions,
    setDescription,
    setHasUnlimitedParticipants,
    setIsPaid,
    setIsPrivate,
    setName,
    setParticipantLimit,
    setPaymentAmount,
    successMessage,
    transferringOwnerUserID,
    updatingPaymentUserID,
    handleApprove,
    handleRemoveMember,
    handleSaveGroup,
    handleTransferOwnership,
    handleUpdatePayment,
  } = useGroupAdminScreen(group, onGroupUpdated, onBack);

  function confirmRemoveMember(member: (typeof members)[number]) {
    const name = member.display_name || `Usuário ${member.user_id.slice(0, 8)}`;

    Alert.alert(
      'Remover participante',
      `Você tem certeza que deseja remover ${name} deste grupo?`,
      [
        { style: 'cancel', text: 'Cancelar' },
        { onPress: () => handleRemoveMember(member), style: 'destructive', text: 'Remover' },
      ],
    );
  }

  function confirmTransferOwnership(member: (typeof members)[number]) {
    const name = member.display_name || `Usuário ${member.user_id.slice(0, 8)}`;

    Alert.alert(
      'Transferir propriedade',
      `Você tem certeza que deseja tornar ${name} dono deste grupo? Você deixará de administrar o grupo.`,
      [
        { style: 'cancel', text: 'Cancelar' },
        { onPress: () => handleTransferOwnership(member), text: 'Transferir' },
      ],
    );
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <ScrollView contentContainerStyle={styles.container} showsVerticalScrollIndicator={false}>
        <View style={styles.backgroundMarkerTop} />
        <View style={styles.backgroundCircle} />

        <BackButton onPress={onBack} />

        <GroupAdminHeader groupName={group.name} error={error} successMessage={successMessage} />

        <GroupAdminForm
          blockPendingPredictions={blockPendingPredictions}
          description={description}
          hasUnlimitedParticipants={hasUnlimitedParticipants}
          isPaid={isPaid}
          isPrivate={isPrivate}
          isSaving={isSaving}
          name={name}
          participantLimit={participantLimit}
          paymentAmount={paymentAmount}
          onSave={handleSaveGroup}
          setBlockPendingPredictions={setBlockPendingPredictions}
          setDescription={setDescription}
          setHasUnlimitedParticipants={setHasUnlimitedParticipants}
          setIsPaid={setIsPaid}
          setIsPrivate={setIsPrivate}
          setName={setName}
          setParticipantLimit={setParticipantLimit}
          setPaymentAmount={setPaymentAmount}
        />

        {group.role === 'owner' && group.is_paid ? (
          <GroupAdminPayments
            isLoadingPayments={isLoadingPayments}
            loadPayments={loadPayments}
            onUpdatePayment={handleUpdatePayment}
            payments={payments}
            summary={paymentsSummary}
            updatingPaymentUserID={updatingPaymentUserID}
          />
        ) : null}

        <GroupAdminRequests
          approvingUserID={approvingUserID}
          isLoadingRequests={isLoadingRequests}
          loadRequests={loadRequests}
          onApprove={handleApprove}
          requests={requests}
        />

        <GroupAdminMembers
          isLoadingMembers={isLoadingMembers}
          loadMembers={loadMembers}
          members={members}
          onRemove={confirmRemoveMember}
          onTransferOwnership={confirmTransferOwnership}
          removingUserID={removingUserID}
          transferringOwnerUserID={transferringOwnerUserID}
        />
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
  backgroundMarkerTop: {
    borderColor: 'rgba(255, 255, 255, 0.68)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  backgroundCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 140,
    borderWidth: 2,
    height: 280,
    position: 'absolute',
    right: -128,
    top: 104,
    width: 280,
  },
});
