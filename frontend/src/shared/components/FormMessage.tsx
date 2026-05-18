import { StyleSheet, Text } from 'react-native';

import { colors, typography } from '../theme';

type FormMessageProps = {
  message: string | null;
  tone?: 'error' | 'success';
};

export function FormMessage({ message, tone = 'error' }: FormMessageProps) {
  if (!message) {
    return null;
  }

  return <Text style={[styles.message, tone === 'success' && styles.success]}>{message}</Text>;
}

const styles = StyleSheet.create({
  message: {
    color: colors.dangerStrong,
    fontSize: typography.body,
    fontWeight: '700',
  },
  success: {
    color: colors.primary,
  },
});
