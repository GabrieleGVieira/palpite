import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Header } from '../global/Header';

type HomeHeaderProps = {
  userName?: string;
  onCreateGroup: () => void;
  onLogout: () => void;
  isSubmitting: boolean;
};

export function HomeHeader({ userName, onCreateGroup, onLogout, isSubmitting }: HomeHeaderProps) {
  return (
    <View>
      <Header
        title={`Olá, ${userName || 'amigo'}`}
        subtitle="Aqui você acompanha seus grupos e seus palpites com calma."
      />

      <View style={styles.actionTabs}>
        <Pressable onPress={onCreateGroup} style={[styles.tabButton, styles.tabPrimary]}>
          <Text style={styles.tabButtonText}>Criar grupo</Text>
        </Pressable>
        <Pressable
          disabled={isSubmitting}
          onPress={onLogout}
          style={[styles.tabButton, styles.tabSecondary, isSubmitting && styles.buttonDisabled]}>
          <Text style={styles.tabSecondaryText}>{isSubmitting ? 'Saindo...' : 'Sair'}</Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  actionTabs: {
    flexDirection: 'row',
    gap: 10,
    marginTop: 18,
  },
  tabButton: {
    flex: 1,
    alignItems: 'center',
    borderRadius: 999,
    justifyContent: 'center',
    minHeight: 48,
    paddingHorizontal: 12,
  },
  tabPrimary: {
    backgroundColor: '#1f7a4a',
  },
  tabSecondary: {
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderWidth: 1,
  },
  tabButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '800',
  },
  tabSecondaryText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
  },
  buttonDisabled: {
    opacity: 0.72,
  },
});
