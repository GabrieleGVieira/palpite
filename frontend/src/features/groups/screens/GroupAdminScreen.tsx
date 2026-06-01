import { StatusBar } from 'expo-status-bar';
import { useMemo, useState } from 'react';
import { Alert, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { BackButton } from '../../../shared/components/BackButton';
import { EmptyBox } from '../../../shared/components/EmptyBox';
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

type GroupAdminTab = 'edit' | 'participants' | 'requests';

export function GroupAdminScreen({ group, onBack, onGroupUpdated }: GroupAdminScreenProps) {
  const [activeTab, setActiveTab] = useState<GroupAdminTab>('edit');
  const {
    approvingUserID,
    blockPendingPredictions,
    description,
    hasUnlimitedParticipants,
    isLoadingMembers,
    isLoadingPayments,
    isLoadingRequests,
    isPaid,
    isPaymentControlSaved,
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
    transferringOwnerUserID,
    updatingPaymentUserID,
    handleApprove,
    handleRemoveMember,
    handleSaveGroup,
    handleTransferOwnership,
    handleUpdatePayment,
  } = useGroupAdminScreen(group, onGroupUpdated, onBack);
  const paymentsByUserID = useMemo(
    () =>
      Object.fromEntries(payments.map((payment) => [payment.user_id, payment])),
    [payments],
  );

  function confirmRemoveMember(member: (typeof members)[number]) {
    const name = member.display_name || `Usuário ${member.user_id.slice(0, 8)}`;

    Alert.alert(
      'Remover Palpiteiro',
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

        <GroupAdminHeader groupName={group.name} />

        <View style={styles.tabs}>
          <AdminTabButton
            activeTab={activeTab}
            label="Editar"
            tab="edit"
            onChangeTab={setActiveTab}
          />
          <AdminTabButton
            activeTab={activeTab}
            label="Palpiteiros"
            tab="participants"
            onChangeTab={setActiveTab}
          />
          <AdminTabButton
            activeTab={activeTab}
            label="Solicitações"
            tab="requests"
            onChangeTab={setActiveTab}
          />
        </View>

        {activeTab === 'edit' ? (
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
        ) : null}

        {activeTab === 'participants' ? (
          <>
            {group.role === 'owner' && isPaid && isPaymentControlSaved ? (
              <GroupAdminPayments
                isLoadingPayments={isLoadingPayments}
                loadPayments={loadPayments}
                onUpdatePayment={handleUpdatePayment}
                payments={payments}
                summary={paymentsSummary}
                updatingPaymentUserID={updatingPaymentUserID}
              />
            ) : null}

            {group.role === 'owner' && isPaid && !isPaymentControlSaved ? (
              <EmptyBox
                title="Salve a participação paga."
                text="Depois de salvar o grupo, os controles de pagamento dos Palpiteiros aparecem aqui."
              />
            ) : null}

            <GroupAdminMembers
              isLoadingMembers={isLoadingMembers}
              loadMembers={loadMembers}
              members={members}
              onRemove={confirmRemoveMember}
              onTransferOwnership={confirmTransferOwnership}
              paymentsByUserID={paymentsByUserID}
              removingUserID={removingUserID}
              transferringOwnerUserID={transferringOwnerUserID}
            />
          </>
        ) : null}

        {activeTab === 'requests' ? (
          <GroupAdminRequests
            approvingUserID={approvingUserID}
            isLoadingRequests={isLoadingRequests}
            loadRequests={loadRequests}
            onApprove={handleApprove}
            requests={requests}
          />
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}

function AdminTabButton({
  activeTab,
  label,
  onChangeTab,
  tab,
}: {
  activeTab: GroupAdminTab;
  label: string;
  onChangeTab: (tab: GroupAdminTab) => void;
  tab: GroupAdminTab;
}) {
  const isActive = activeTab === tab;

  return (
    <Pressable
      onPress={() => onChangeTab(tab)}
      style={[styles.tabButton, isActive && styles.tabButtonActive]}>
      <Text style={[styles.tabButtonText, isActive && styles.tabButtonTextActive]}>{label}</Text>
    </Pressable>
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
  tabs: {
    backgroundColor: '#edf3e8',
    borderRadius: 8,
    flexDirection: 'row',
    gap: 6,
    padding: 6,
  },
  tabButton: {
    alignItems: 'center',
    borderRadius: 7,
    flex: 1,
    justifyContent: 'center',
    minHeight: 44,
  },
  tabButtonActive: {
    backgroundColor: '#ffffff',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 4, width: 0 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
  },
  tabButtonText: {
    color: '#486654',
    fontSize: 13,
    fontWeight: '800',
    textAlign: 'center',
  },
  tabButtonTextActive: {
    color: '#1f7a4a',
  },
});
