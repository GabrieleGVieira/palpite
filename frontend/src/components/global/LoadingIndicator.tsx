import { ActivityIndicator, View, Text, StyleSheet } from 'react-native';

export function LoadingIndicator({ text }: { text: string }) {
  return (
    <View style={styles.loadingBox}>
      <ActivityIndicator color="#1f7a4a" />
      <Text style={styles.loadingText}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  loadingBox: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    gap: 8,
    padding: 18,
  },
  loadingText: {
    color: '#486654',
    fontSize: 14,
  },
});
