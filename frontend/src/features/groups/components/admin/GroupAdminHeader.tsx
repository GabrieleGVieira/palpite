import { StyleSheet, Text, View } from 'react-native';

type GroupAdminHeaderProps = {
  groupName: string;
};

export function GroupAdminHeader({ groupName }: GroupAdminHeaderProps) {
  return (
    <View style={styles.headerBlock}>
      <View>
        <Text style={styles.title}>Admin do grupo</Text>
        <Text style={styles.subtitle}>{groupName}</Text>
      </View>
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
});
