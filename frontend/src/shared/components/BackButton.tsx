import { Pressable, Text, StyleSheet } from 'react-native';

export function BackButton({ onPress }: { onPress: () => void }) {
  return (
    <Pressable accessibilityLabel="Voltar" onPress={onPress} style={styles.backButton}>
      <Text style={styles.backButtonText}>‹</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 6, width: 0 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
});
