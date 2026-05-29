import { Pressable, Text, StyleSheet } from 'react-native';

import { colors, typography } from '../theme';

type Props = {
  disabledLabel?: string;
  isDisabled?: boolean;
  isLoading: boolean;
  onPress: () => void;
  loadingLabel: string;
  waitingLabel: string;
};
export function FinishButton({
  disabledLabel,
  isDisabled = false,
  isLoading,
  onPress,
  loadingLabel,
  waitingLabel,
}: Props) {
  const isUnavailable = isDisabled || isLoading;
  const label = isLoading
    ? loadingLabel
    : isDisabled && disabledLabel
      ? disabledLabel
      : waitingLabel;

  return (
    <Pressable
      disabled={isUnavailable}
      onPress={onPress}
      style={[
        styles.primaryButton,
        isDisabled && styles.primaryButtonDisabled,
        isLoading && styles.buttonDisabled,
      ]}>
      <Text style={styles.primaryButtonText}>{label}</Text>
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
    padding: 5
  },
  primaryButtonDisabled: {
    backgroundColor: '#8a9490',
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
