import { Text, TextInput, StyleSheet, View } from 'react-native';
import { SwitchBox } from './SwitchBox';

type ParticipantsCardProps = {
  hasUnlimitedParticipants: boolean;
  participantLimit: string;
  setHasUnlimitedParticipants: (value: boolean) => void;
  setParticipantLimit: (value: string) => void;
};

export function ParticipantsCard({
  hasUnlimitedParticipants,
  participantLimit,
  setHasUnlimitedParticipants,
  setParticipantLimit,
}: ParticipantsCardProps) {
  return (
    <View style={styles.row}>
      <View style={[styles.fieldGroup, styles.limitField]}>
        <Text style={styles.label}>Palpiteiros</Text>
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

      <SwitchBox
        title="Ilimitado"
        subtitle="Sem teto"
        value={hasUnlimitedParticipants}
        onPress={setHasUnlimitedParticipants}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  fieldGroup: {
    gap: 8,
  },
  row: {
    flexDirection: 'row',
    gap: 12,
  },
  limitField: {
    flex: 1,
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
});
