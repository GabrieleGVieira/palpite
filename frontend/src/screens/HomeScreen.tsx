import { StatusBar } from 'expo-status-bar';
import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth } from '../hooks/useAuth';

type HomeScreenProps = {
  onCreateGroup: () => void;
};

export function HomeScreen({ onCreateGroup }: HomeScreenProps) {
  const { isSubmitting, logout, user } = useAuth();
  const userName = user?.user_metadata.full_name as string | undefined;

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <View style={styles.container}>
        <View style={styles.pitchMarkTop} />
        <View style={styles.pitchCircle} />

        <View style={styles.header}>
          <View style={styles.logoMark}>
            <Image
              accessibilityIgnoresInvertColors
              resizeMode="cover"
              source={require('../../assets/splash-palpitai.png')}
              style={styles.logoImage}
            />
          </View>
          <Text style={styles.title}>PalpitAI</Text>
          <Text style={styles.subtitle}>
            Sessao ativa via Supabase Auth. Seus grupos, palpites e ranking entram aqui.
          </Text>
        </View>

        <View style={styles.sessionBox}>
          <Text style={styles.sessionLabel}>Conta conectada</Text>
          <Text style={styles.sessionName}>{userName || user?.email || 'Usuario'}</Text>
          {user?.email ? <Text style={styles.sessionEmail}>{user.email}</Text> : null}
        </View>

        <View style={styles.actions}>
          <Pressable onPress={onCreateGroup} style={styles.primaryButton}>
            <Text style={styles.primaryButtonText}>Criar grupo</Text>
          </Pressable>

          <Pressable disabled={isSubmitting} onPress={logout} style={styles.secondaryButton}>
            <Text style={styles.secondaryButtonText}>{isSubmitting ? 'Saindo...' : 'Sair'}</Text>
          </Pressable>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  container: {
    flex: 1,
    backgroundColor: '#f5f8ef',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    paddingVertical: 32,
  },
  pitchMarkTop: {
    borderColor: 'rgba(255, 255, 255, 0.68)',
    borderRadius: 8,
    borderWidth: 2,
    height: 116,
    left: 24,
    position: 'absolute',
    right: 24,
    top: -42,
  },
  pitchCircle: {
    borderColor: 'rgba(32, 111, 67, 0.12)',
    borderRadius: 140,
    borderWidth: 2,
    height: 280,
    position: 'absolute',
    right: -128,
    top: 104,
    width: 280,
  },
  header: {
    paddingTop: 32,
  },
  logoMark: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 32,
    borderWidth: 1,
    height: 64,
    justifyContent: 'center',
    marginBottom: 24,
    overflow: 'hidden',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 8, width: 0 },
    shadowOpacity: 0.12,
    shadowRadius: 16,
    width: 64,
  },
  logoImage: {
    height: 76,
    transform: [{ scale: 1.18 }],
    width: 76,
  },
  title: {
    color: '#123d2a',
    fontSize: 38,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 340,
  },
  sessionBox: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    padding: 20,
  },
  sessionLabel: {
    color: '#486654',
    fontSize: 13,
    fontWeight: '700',
    textTransform: 'uppercase',
  },
  sessionName: {
    color: '#123d2a',
    fontSize: 24,
    fontWeight: '800',
    marginTop: 8,
  },
  sessionEmail: {
    color: '#486654',
    fontSize: 15,
    marginTop: 4,
  },
  actions: {
    gap: 12,
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    minHeight: 54,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '800',
  },
  secondaryButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#1f7a4a',
    borderRadius: 8,
    borderWidth: 1,
    justifyContent: 'center',
    minHeight: 54,
  },
  secondaryButtonText: {
    color: '#1f7a4a',
    fontSize: 16,
    fontWeight: '800',
  },
});
