import { StatusBar } from 'expo-status-bar';
import {
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { BackButton } from '../../../shared/components/BackButton';
import { useCreateGroupScreen } from '../hooks/useCreateGroupScreen';
import { GroupBasicFormFields } from '../../../shared/components/GroupBasicFormFields';
import { CreateGroupMatchScope } from '../components/create-group/CreateGroupMatchScope';
import { ParticipantsCard } from '../../../shared/components/ParticipantsCard';
import { SwitchBox } from '../../../shared/components/SwitchBox';
import { Header } from '../../../shared/components/Header';
import { PaymentSettingsCard } from '../../../shared/components/PaymentSettingsCard';

type CreateGroupScreenProps = {
  onBack: () => void;
  onGroupCreated: () => void;
};

export function CreateGroupScreen({ onBack, onGroupCreated }: CreateGroupScreenProps) {
  const {
    blockPendingPredictions,
    description,
    hasUnlimitedParticipants,
    isPaid,
    isPrivate,
    isSubmitting,
    matchScope,
    matchScopes,
    participantLimit,
    paymentAmount,
    selectedTeams,
    teamSearch,
    toggleTeamDropdown,
    isTeamDropdownOpen,
    filteredTeams,
    groupName,
    onChangeDescription,
    onChangeGroupName,
    onChangeMatchScope,
    onChangeParticipantLimit,
    onChangeTeamSearch,
    onCreateGroup,
    toggleTeam,
    setBlockPendingPredictions,
    setHasUnlimitedParticipants,
    setIsPaid,
    setIsPrivate,
    setPaymentAmount,
  } = useCreateGroupScreen(onGroupCreated, onBack);

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}>
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}>
          <View style={styles.pitchMarkTop} />
          <View style={styles.pitchCircle} />

          <BackButton onPress={onBack} />

          <Header
            title="Criar grupo"
            subtitle="Monte um bolão da Copa do Mundo, defina quais jogos entram e convide sua turma para disputar o ranking."
            withImage={false}
          />

          <View style={styles.form}>
            <GroupBasicFormFields
              description={description}
              groupName={groupName}
              onChangeDescription={onChangeDescription}
              onChangeGroupName={onChangeGroupName}
            />

            <CreateGroupMatchScope
              matchScope={matchScope}
              matchScopes={matchScopes}
              onChangeMatchScope={onChangeMatchScope}
              filteredTeams={filteredTeams}
              isOpen={isTeamDropdownOpen}
              onChangeSearchText={onChangeTeamSearch}
              onToggleOpen={toggleTeamDropdown}
              onToggleTeam={toggleTeam}
              searchText={teamSearch}
              selectedTeams={selectedTeams}
            />

            <ParticipantsCard
              hasUnlimitedParticipants={hasUnlimitedParticipants}
              participantLimit={participantLimit}
              setHasUnlimitedParticipants={setHasUnlimitedParticipants}
              setParticipantLimit={onChangeParticipantLimit}
            />

            <PaymentSettingsCard
              blockPendingPredictions={blockPendingPredictions}
              isPaid={isPaid}
              paymentAmount={paymentAmount}
              setBlockPendingPredictions={setBlockPendingPredictions}
              setIsPaid={setIsPaid}
              setPaymentAmount={setPaymentAmount}
            />

            <SwitchBox
              title="Privado"
              subtitle="Novos membros precisam de aprovação"
              value={isPrivate}
              onPress={setIsPrivate}
            />

            <View style={styles.rulesBox}>
              <Text style={styles.rulesTitle}>Regras iniciais</Text>
              <Text style={styles.rulesText}>• 10 pontos por placar exato</Text>
              <Text style={styles.rulesText}>• 7 pontos por vencedor/empate e quantidade de gols de um dos times corretos</Text>
              <Text style={styles.rulesText}>• 5 pontos por vencedor ou empate correto</Text>
              <Text style={styles.rulesText}>• 3 pontos por quantidade de gols de um dos times mas sem acertar o vencedor</Text>
              <Text style={styles.rulesText}>• Palpites fecham no início de cada jogo</Text>
            </View>

            <Pressable
              disabled={isSubmitting}
              onPress={onCreateGroup}
              style={[styles.primaryButton, isSubmitting && styles.primaryButtonDisabled]}>
              <Text style={styles.primaryButtonText}>
                {isSubmitting ? 'Criando grupo...' : 'Criar grupo'}
              </Text>
            </Pressable>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  form: {
    gap: 16,
    marginTop: 36,
  },
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  keyboardView: {
    flex: 1,
  },
  scrollContent: {
    backgroundColor: '#f5f8ef',
    flexGrow: 1,
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
  rulesBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 16,
  },
  rulesTitle: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
    marginBottom: 10,
  },
  rulesText: {
    color: '#486654',
    fontSize: 13,
    lineHeight: 20,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1B5B38',
    borderRadius: 12,
    justifyContent: 'center',
    minHeight: 56,
  },
  primaryButtonDisabled: {
    backgroundColor: '#8ca789',
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '700',
  },
});
