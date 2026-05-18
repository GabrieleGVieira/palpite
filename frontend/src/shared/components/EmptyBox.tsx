import { View, Text, StyleSheet } from 'react-native';

export function EmptyBox({ title, text }: { title: string; text: string }) {
  return (
    <View style={styles.emptyBox}>
      <Text style={styles.emptyTitle}>{title}</Text>
      <Text style={styles.emptyText}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  emptyBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 18,
  },
  emptyTitle: {
    color: '#123d2a',
    fontSize: 17,
    fontWeight: '800',
  },
  emptyText: {
    color: '#486654',
    fontSize: 14,
    lineHeight: 20,
    marginTop: 6,
  },
});
