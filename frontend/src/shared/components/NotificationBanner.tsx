import { StyleSheet, Text, View } from 'react-native';

type NotificationBannerProps = {
  message: string | null;
};

export function NotificationBanner({ message }: NotificationBannerProps) {
  if (!message) {
    return null;
  }

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Atualizacão ao vivo</Text>
      <Text style={styles.message}>{message}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#1E1E1E',
    borderColor: '#a03222',
    borderRadius: 8,
    borderWidth: 1,
    padding: 14,
  },
  label: {
    color: '#cde4c9',
    fontSize: 11,
    fontWeight: '900',
    textTransform: 'uppercase',
  },
  message: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '800',
    lineHeight: 20,
    marginTop: 4,
  },
});
