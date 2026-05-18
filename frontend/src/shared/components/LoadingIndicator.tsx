import { ActivityIndicator, View, Text, StyleSheet } from 'react-native';

import { colors, spacing, typography } from '../theme';

export function LoadingIndicator({ text }: { text: string }) {
  return (
    <View style={styles.loadingBox}>
      <ActivityIndicator color={colors.primary} />
      <Text style={styles.loadingText}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  loadingBox: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    gap: spacing.sm,
    padding: 18,
  },
  loadingText: {
    color: colors.mutedText,
    fontSize: typography.body,
  },
});
