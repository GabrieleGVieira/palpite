import { StyleSheet, Text, TextInput, View } from 'react-native';

import { colors, spacing, typography } from '../theme';

type AuthInputFieldProps = {
  label: string;
  placeholder: string;
  secureTextEntry?: boolean;
  value: string;
  onChangeText: (text: string) => void;
  keyboardType?: 'default' | 'email-address';
  autoCapitalize?: 'none' | 'sentences' | 'words' | 'characters';
};

export function AuthInputField({
  label,
  placeholder,
  secureTextEntry = false,
  value,
  onChangeText,
  keyboardType = 'default',
  autoCapitalize = 'none',
}: AuthInputFieldProps) {
  return (
    <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <TextInput
        autoCapitalize={autoCapitalize}
        keyboardType={keyboardType}
        onChangeText={onChangeText}
        placeholder={placeholder}
        placeholderTextColor="#9ca6a0"
        secureTextEntry={secureTextEntry}
        selectionColor={colors.textSecondary}
        style={styles.input}
        value={value}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  field: {
    marginBottom: spacing.xl,
  },
  label: {
    color: colors.textSecondary,
    fontSize: typography.label,
    fontWeight: '700',
    letterSpacing: 0.2,
    marginBottom: spacing.sm,
  },
  input: {
    backgroundColor: colors.fieldBackground,
    borderColor: colors.fieldBorder,
    borderRadius: 20,
    borderWidth: 1,
    color: colors.primaryText,
    fontSize: typography.input,
    minHeight: 50,
    paddingHorizontal: 18,
    paddingVertical: 14,
  },
});
