import { StyleSheet, Text, TextInput, View } from 'react-native';
import { AuthInputField } from './AuthInputField';

type GroupBasicFormFieldsProps = {
  description: string;
  groupName: string;
  onChangeDescription: (value: string) => void;
  onChangeGroupName: (value: string) => void;
};

export function GroupBasicFormFields({
  description,
  groupName,
  onChangeDescription,
  onChangeGroupName,
}: GroupBasicFormFieldsProps) {
  return (
    <View style={styles.container}>
      <AuthInputField
        autoCapitalize="words"
        keyboardType="default"
        label="Nome do grupo"
        onChangeText={onChangeGroupName}
        placeholder="Ex: Família na Copa"
        value={groupName}
      />

      <View style={styles.fieldGroup}>
        <Text style={styles.label}>Descrição</Text>
        <TextInput
          multiline
          onChangeText={onChangeDescription}
          placeholder="Regras, combinados ou prêmio simbólico"
          placeholderTextColor="#7c8898"
          style={[styles.input, styles.textArea]}
          textAlignVertical="top"
          value={description}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    gap: 16,
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
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 16,
    minHeight: 54,
    paddingHorizontal: 16,
  },
  textArea: {
    minHeight: 104,
    paddingTop: 14,
  },
});
