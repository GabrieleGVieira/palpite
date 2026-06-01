import { Image, StyleSheet, Text, View } from 'react-native';

import { colors } from '../../../shared/theme';

type Props = {
  name: string;
  uri?: string | null;
  size?: number;
};

export function UserAvatar({ name, uri, size = 48 }: Props) {
  const style = { borderRadius: size / 2, height: size, width: size };

  if (uri?.trim()) {
    return <Image source={{ uri: uri.trim() }} style={[styles.avatar, style]} />;
  }

  return (
    <View style={[styles.avatarFallback, style]}>
      <Text style={styles.avatarText}>{initials(name)}</Text>
    </View>
  );
}

function initials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || '?';
}

const styles = StyleSheet.create({
  avatar: {
    backgroundColor: colors.fieldBackground,
  },
  avatarFallback: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    justifyContent: 'center',
  },
  avatarText: {
    color: colors.primary,
    fontSize: 15,
    fontWeight: '800',
  },
});
