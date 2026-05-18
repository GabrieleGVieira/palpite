import { StyleSheet, Text, View } from 'react-native';

import { SwitchBox } from '../../../../shared/components/SwitchBox';
import { ParticipantsCard } from '../../../../shared/components/ParticipantsCard';
import { GroupBasicFormFields } from '../../../../shared/components/GroupBasicFormFields';
import { FinishButton } from '../../../../shared/components/FinishButton';

type GroupAdminFormProps = {
  description: string;
  hasUnlimitedParticipants: boolean;
  isPrivate: boolean;
  isSaving: boolean;
  name: string;
  participantLimit: string;
  onSave: () => void;
  setDescription: (value: string) => void;
  setHasUnlimitedParticipants: (value: boolean) => void;
  setIsPrivate: (value: boolean) => void;
  setName: (value: string) => void;
  setParticipantLimit: (value: string) => void;
};

export function GroupAdminForm({
  description,
  hasUnlimitedParticipants,
  isPrivate,
  isSaving,
  name,
  participantLimit,
  onSave,
  setDescription,
  setHasUnlimitedParticipants,
  setIsPrivate,
  setName,
  setParticipantLimit,
}: GroupAdminFormProps) {
  return (
    <View style={styles.card}>
      <Text style={styles.cardTitle}>Informações</Text>

      <GroupBasicFormFields
        description={description}
        groupName={name}
        onChangeDescription={setDescription}
        onChangeGroupName={setName}
      />

      <ParticipantsCard
        hasUnlimitedParticipants={hasUnlimitedParticipants}
        participantLimit={participantLimit}
        setHasUnlimitedParticipants={setHasUnlimitedParticipants}
        setParticipantLimit={setParticipantLimit}
      />

      <SwitchBox
        title="Privado"
        subtitle="Novos membros precisam de aprovação"
        value={isPrivate}
        onPress={setIsPrivate}
      />

      <FinishButton
        isLoading={isSaving}
        onPress={onSave}
        loadingLabel="Salvando..."
        waitingLabel="Salvar"
      />
    </View>
  );
}

const styles = StyleSheet.create({
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
});
