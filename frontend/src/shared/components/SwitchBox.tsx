import { View, Text, StyleSheet, Switch } from 'react-native';

export function SwitchBox({
  title,
  subtitle,
  value,
  onPress,
}: {
  title: string;
  subtitle: string;
  value: boolean;
  onPress: (newValue: boolean) => void;
}) {
  return (
    <View style={styles.switchBoxWide}>
      <View>
        <Text style={styles.switchTitle}>{title}</Text>
        <Text style={styles.switchSubtitle}>{subtitle}</Text>
      </View>
      <Switch
        onValueChange={onPress}
        thumbColor="#ffffff"
        trackColor={{ false: '#c9d7c3', true: '#1f7a4a' }}
        value={value}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  switchBox: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flex: 1.1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 78,
    paddingHorizontal: 12,
  },
  switchBoxWide: {
    alignItems: 'center',
    backgroundColor: '#f5f8ef',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 78,
    paddingHorizontal: 12,
  },
  switchTitle: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '800',
  },
  switchSubtitle: {
    color: '#486654',
    fontSize: 12,
    marginTop: 4,
  },
  toggleButton: {
    backgroundColor: '#1f7a4a',
    borderRadius: 16,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  toggleText: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '700',
  },
});
