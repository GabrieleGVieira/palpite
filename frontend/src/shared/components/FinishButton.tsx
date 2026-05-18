import { Pressable, Text, StyleSheet } from 'react-native';

import { colors, typography } from '../theme';

type Props = {
  isLoading: boolean;
  onPress: () => void;
  loadingLabel: string;
  waitingLabel: string;
};
export function FinishButton({ isLoading, onPress, loadingLabel, waitingLabel }: Props) {
  return (
    <Pressable
      disabled={isLoading}
      onPress={onPress}
      style={[styles.primaryButton, isLoading && styles.buttonDisabled]}>
      <Text style={styles.primaryButtonText}>{isLoading ? loadingLabel : waitingLabel}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 52,
  },
  primaryButtonText: {
    color: colors.white,
    fontSize: typography.button,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.72,
  },
});
