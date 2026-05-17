import { StatusBar } from 'expo-status-bar';
import { useState } from 'react';
import {
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth } from '../hooks/useAuth';

type SignupScreenProps = {
  onBackToLogin: () => void;
};

export function SignupScreen({ onBackToLogin }: SignupScreenProps) {
  const { clearError, error, isConfigured, isSubmitting, signup } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function handleSignup() {
    clearError();
    setFormError(null);
    setSuccessMessage(null);

    if (!name.trim() || !email.trim() || !password || !confirmPassword) {
      setFormError('Preencha todos os campos para criar sua conta.');
      return;
    }

    if (password !== confirmPassword) {
      setFormError('As senhas precisam ser iguais.');
      return;
    }

    const created = await signup(name, email, password);

    if (created) {
      onBackToLogin();
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}>
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}>
          <View style={styles.pitchMarkTop} />
          <View style={styles.pitchCircle} />

          <Pressable
            accessibilityLabel="Voltar para o login"
            onPress={onBackToLogin}
            style={styles.backButton}>
            <Text style={styles.backButtonText}>‹</Text>
          </Pressable>

          <View style={styles.header}>
            <View style={styles.logoMark}>
              <Image
                accessibilityIgnoresInvertColors
                resizeMode="cover"
                source={require('../../assets/splash-palpitai.png')}
                style={styles.logoImage}
              />
            </View>
            <Text style={styles.title}>Crie sua conta</Text>
            <Text style={styles.subtitle}>
              Monte seus grupos, registre palpites e acompanhe os insights do PalpitAI durante a
              competicao.
            </Text>
          </View>

          <View style={styles.form}>
            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Nome</Text>
              <TextInput
                autoCapitalize="words"
                autoComplete="name"
                onChangeText={setName}
                placeholder="Seu nome"
                placeholderTextColor="#7c8898"
                style={styles.input}
                textContentType="name"
                value={name}
              />
            </View>

            <View style={styles.fieldGroup}>
              <Text style={styles.label}>E-mail</Text>
              <TextInput
                autoCapitalize="none"
                autoComplete="email"
                keyboardType="email-address"
                onChangeText={setEmail}
                placeholder="voce@email.com"
                placeholderTextColor="#7c8898"
                style={styles.input}
                textContentType="emailAddress"
                value={email}
              />
            </View>

            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Senha</Text>
              <TextInput
                autoCapitalize="none"
                autoComplete="new-password"
                onChangeText={setPassword}
                placeholder="Crie uma senha"
                placeholderTextColor="#7c8898"
                secureTextEntry
                style={styles.input}
                textContentType="newPassword"
                value={password}
              />
            </View>

            <View style={styles.fieldGroup}>
              <Text style={styles.label}>Confirmar senha</Text>
              <TextInput
                autoCapitalize="none"
                autoComplete="new-password"
                onChangeText={setConfirmPassword}
                placeholder="Repita sua senha"
                placeholderTextColor="#7c8898"
                secureTextEntry
                style={styles.input}
                textContentType="newPassword"
                value={confirmPassword}
              />
            </View>

            {!isConfigured ? (
              <Text style={styles.feedbackText}>
                Configure EXPO_PUBLIC_SUPABASE_URL e EXPO_PUBLIC_SUPABASE_KEY.
              </Text>
            ) : null}

            {formError ? <Text style={styles.errorText}>{formError}</Text> : null}
            {error ? <Text style={styles.errorText}>{error}</Text> : null}
            {successMessage ? <Text style={styles.successText}>{successMessage}</Text> : null}

            <Pressable
              disabled={isSubmitting || !isConfigured}
              onPress={handleSignup}
              style={styles.primaryButton}>
              <Text style={styles.primaryButtonText}>
                {isSubmitting ? 'Criando conta...' : 'Criar conta'}
              </Text>
            </Pressable>
          </View>

          <View style={styles.footer}>
            <Text style={styles.footerText}>Já tem conta?</Text>
            <Pressable onPress={onBackToLogin}>
              <Text style={styles.footerLink}> Entrar</Text>
            </Pressable>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: '#f5f8ef',
  },
  keyboardView: {
    flex: 1,
  },
  scrollContent: {
    backgroundColor: '#f5f8ef',
    flexGrow: 1,
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
  backButton: {
    alignItems: 'center',
    backgroundColor: '#ffffff',
    borderColor: '#d9e7d4',
    borderRadius: 22,
    borderWidth: 1,
    height: 44,
    justifyContent: 'center',
    shadowColor: '#1e5c39',
    shadowOffset: { height: 6, width: 0 },
    shadowOpacity: 0.1,
    shadowRadius: 12,
    width: 44,
  },
  backButtonText: {
    color: '#1f7a4a',
    fontSize: 34,
    fontWeight: '600',
    lineHeight: 38,
  },
  header: {
    paddingTop: 20,
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
    fontSize: 36,
    fontWeight: '800',
    letterSpacing: 0,
  },
  subtitle: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 360,
  },
  form: {
    gap: 16,
    marginTop: 40,
  },
  fieldGroup: {
    gap: 8,
  },
  label: {
    color: '#183f2d',
    fontSize: 14,
    fontWeight: '700',
  },
  input: {
    backgroundColor: '#ffffff',
    borderColor: '#cfe0c9',
    borderRadius: 8,
    borderWidth: 1,
    color: '#183f2d',
    fontSize: 16,
    minHeight: 54,
    paddingHorizontal: 16,
  },
  feedbackText: {
    color: '#8a5f00',
    fontSize: 13,
    lineHeight: 18,
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
  primaryButton: {
    alignItems: 'center',
    backgroundColor: '#1f7a4a',
    borderRadius: 8,
    justifyContent: 'center',
    marginTop: 8,
    minHeight: 54,
  },
  primaryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '800',
  },
  footer: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'center',
    paddingBottom: 8,
    paddingTop: 32,
  },
  footerText: {
    color: '#486654',
    fontSize: 14,
  },
  footerLink: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '800',
  },
});
