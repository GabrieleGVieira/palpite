import { Image, Modal, Pressable, StyleSheet, Text, View } from 'react-native';
import { useState } from 'react';

import { colors } from '../../../../shared/theme';

type HomeHeaderProps = {
  avatarURL?: string | null;
  userName?: string;
  onCreateGroup: () => void;
  onDeleteAccount: () => void;
  onLogout: () => void;
  onOpenProfile: () => void;
  onOpenPrivacy: () => void;
  isSubmitting: boolean;
};

export function HomeHeader({
  avatarURL,
  userName,
  onCreateGroup,
  onDeleteAccount,
  onLogout,
  onOpenProfile,
  onOpenPrivacy,
  isSubmitting,
}: HomeHeaderProps) {
  const [isMenuVisible, setIsMenuVisible] = useState(false);
  const displayName = userName?.trim() || 'amigo';

  function closeMenuAndRun(action: () => void) {
    setIsMenuVisible(false);
    action();
  }

  return (
    <View style={styles.header}>
      <View style={styles.topBar}>
        <View style={styles.logoMark}>
          <Image
            accessibilityIgnoresInvertColors
            resizeMode="cover"
            source={require('../../../../../assets/splash-palpite.png')}
            style={styles.logoImage}
          />
        </View>

        <Pressable
          accessibilityLabel="Abrir menu do perfil"
          onPress={() => setIsMenuVisible(true)}
          style={({ pressed }) => [styles.avatarButton, pressed && styles.pressed]}>
          {avatarURL?.trim() ? (
            <Image source={{ uri: avatarURL.trim() }} style={styles.avatarImage} />
          ) : (
            <Text style={styles.avatarText}>{initials(displayName)}</Text>
          )}
        </Pressable>
      </View>

      <View style={styles.titleBlock}>
        <Text style={styles.title}>Olá, {displayName}</Text>
        <Text style={styles.subtitle}>Acompanhe seus grupos, pontuação e palpites.</Text>
      </View>

      <View style={styles.actionTabs}>
        <Pressable
          onPress={onCreateGroup}
          style={({ pressed }) => [styles.tabButton, pressed && styles.pressed]}>
          <Text style={styles.tabButtonText}>Criar grupo</Text>
        </Pressable>
      </View>

      <Modal
        animationType="fade"
        onRequestClose={() => setIsMenuVisible(false)}
        transparent
        visible={isMenuVisible}>
        <Pressable style={styles.menuBackdrop} onPress={() => setIsMenuVisible(false)}>
          <Pressable style={styles.menu} onPress={(event) => event.stopPropagation()}>
            <Text style={styles.menuTitle}>Minha conta</Text>
            <MenuItem label="Perfil" onPress={() => closeMenuAndRun(onOpenProfile)} />
            <MenuItem
              label="Políticas e privacidade"
              onPress={() => closeMenuAndRun(onOpenPrivacy)}
            />
            <MenuItem
              disabled={isSubmitting}
              label={isSubmitting ? 'Saindo...' : 'Sair'}
              onPress={() => closeMenuAndRun(onLogout)}
            />
            <View style={styles.menuDivider} />
            <MenuItem
              danger
              label="Excluir conta"
              onPress={() => closeMenuAndRun(onDeleteAccount)}
            />
          </Pressable>
        </Pressable>
      </Modal>
    </View>
  );
}

type MenuItemProps = {
  danger?: boolean;
  disabled?: boolean;
  label: string;
  onPress: () => void;
};

function MenuItem({ danger, disabled, label, onPress }: MenuItemProps) {
  return (
    <Pressable
      disabled={disabled}
      onPress={onPress}
      style={({ pressed }) => [
        styles.menuItem,
        pressed && styles.menuItemPressed,
        disabled && styles.disabled,
      ]}>
      <Text style={[styles.menuItemText, danger && styles.menuItemDanger]}>{label}</Text>
    </Pressable>
  );
}

function initials(name: string) {
  const trimmed = name.trim();
  if (!trimmed) {
    return '?';
  }

  return trimmed
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase();
}

const styles = StyleSheet.create({
  header: {
    gap: 18,
  },
  topBar: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  logoMark: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 24,
    borderWidth: 1,
    height: 48,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 48,
  },
  logoImage: {
    height: 58,
    transform: [{ scale: 1.14 }],
    width: 58,
  },
  avatarButton: {
    alignItems: 'center',
    backgroundColor: '#e3efe0',
    borderColor: colors.surface,
    borderRadius: 24,
    borderWidth: 2,
    height: 48,
    justifyContent: 'center',
    overflow: 'hidden',
    width: 48,
  },
  avatarImage: {
    height: 48,
    width: 48,
  },
  avatarText: {
    color: colors.primary,
    fontSize: 16,
    fontWeight: '800',
  },
  titleBlock: {
    gap: 6,
  },
  title: {
    color: colors.primaryText,
    fontSize: 30,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: colors.mutedText,
    fontSize: 15,
    lineHeight: 22,
  },
  actionTabs: {
    flexDirection: 'row',
  },
  tabButton: {
    alignItems: 'center',
    backgroundColor: colors.primary,
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 46,
    paddingHorizontal: 18,
  },
  tabButtonText: {
    color: colors.white,
    fontSize: 14,
    fontWeight: '800',
  },
  menuBackdrop: {
    flex: 1,
  },
  menu: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: 8,
    borderWidth: 1,
    elevation: 8,
    minWidth: 236,
    padding: 8,
    position: 'absolute',
    right: 24,
    shadowColor: '#163f2c',
    shadowOffset: { height: 10, width: 0 },
    shadowOpacity: 0.16,
    shadowRadius: 18,
    top: 72,
  },
  menuTitle: {
    color: colors.mutedText,
    fontSize: 12,
    fontWeight: '800',
    paddingHorizontal: 10,
    paddingVertical: 8,
    textTransform: 'uppercase',
  },
  menuItem: {
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 42,
    paddingHorizontal: 10,
  },
  menuItemPressed: {
    backgroundColor: colors.fieldBackground,
  },
  menuItemText: {
    color: colors.primaryText,
    fontSize: 15,
    fontWeight: '700',
  },
  menuItemDanger: {
    color: colors.dangerStrong,
  },
  menuDivider: {
    backgroundColor: colors.border,
    height: 1,
    marginVertical: 6,
  },
  disabled: {
    opacity: 0.72,
  },
  pressed: {
    opacity: 0.86,
  },
});
