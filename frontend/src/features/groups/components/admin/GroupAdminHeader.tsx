import { StyleSheet, Text, View } from 'react-native';

type GroupAdminHeaderProps = {
  groupName: string;
  error: string | null;
  successMessage: string | null;
};

export function GroupAdminHeader({ groupName, error, successMessage }: GroupAdminHeaderProps) {
  return (
    <View style={styles.headerBlock}>
      <View>
        <Text style={styles.title}>Admin do grupo</Text>
        <Text style={styles.subtitle}>{groupName}</Text>
      </View>

      {error ? <Text style={styles.errorText}>{error}</Text> : null}
      {successMessage ? <Text style={styles.successText}>{successMessage}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  headerBlock: {
    gap: 10,
  },
  title: {
    color: '#123d2a',
    fontSize: 34,
    fontWeight: '800',
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    marginTop: 8,
  },
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
  successText: {
    color: '#1f7a4a',
    fontSize: 13,
    lineHeight: 18,
  },
});
