import { StatusBar } from 'expo-status-bar';
import { useState } from 'react';
import {
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth } from '../hooks/useAuth';

type LoginScreenProps = {
  onCreateAccount: () => void;
};

export function LoginScreen({ onCreateAccount }: LoginScreenProps) {
  const { clearError, error, isConfigured, isSubmitting, login, recoverPassword } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function handleLogin() {
    clearError();
    setFormError(null);
    setSuccessMessage(null);

    if (!email.trim() || !password) {
      setFormError('Preencha e-mail e senha para entrar.');
      return;
    }

    await login(email, password);
  }

  async function handleRecoverPassword() {
    clearError();
    setFormError(null);
    setSuccessMessage(null);

    if (!email.trim()) {
      setFormError('Informe seu e-mail para recuperar a senha.');
      return;
    }

    const sent = await recoverPassword(email);

    if (sent) {
      setSuccessMessage('Enviamos as instruções de recuperação para seu e-mail.');
    }
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}>
        <View style={styles.container}>
          <View style={styles.pitchMarkTop} />
          <View style={styles.pitchCircle} />

          <View style={styles.brandBlock}>
            <View style={styles.logoMark}>
              <Image
                accessibilityIgnoresInvertColors
                resizeMode="cover"
                source={require('../../assets/splash-palpitai.png')}
                style={styles.logoImage}
              />
            </View>
            <Text style={styles.appName}>PalpitAI</Text>
            <Text style={styles.tagline}>
              Entre no seu bolão inteligente e acompanhe palpites, rankings e insights com IA.
            </Text>
          </View>

          <View style={styles.form}>
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
                autoComplete="password"
                onChangeText={setPassword}
                placeholder="Sua senha"
                placeholderTextColor="#7c8898"
                secureTextEntry
                style={styles.input}
                textContentType="password"
                value={password}
              />
            </View>

            <Pressable onPress={handleRecoverPassword} style={styles.forgotButton}>
              <Text style={styles.forgotText}>Esqueci minha senha</Text>
            </Pressable>

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
              onPress={handleLogin}
              style={styles.primaryButton}>
              <Text style={styles.primaryButtonText}>
                {isSubmitting ? 'Entrando...' : 'Entrar'}
              </Text>
            </Pressable>
          </View>

          <View style={styles.footer}>
            <Text style={styles.footerText}>Ainda não tem conta?</Text>
            <Pressable onPress={onCreateAccount}>
              <Text style={styles.footerLink}> Criar conta</Text>
            </Pressable>
          </View>
        </View>
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
  brandBlock: {
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
  appName: {
    color: '#123d2a',
    fontSize: 38,
    fontWeight: '800',
    letterSpacing: 0,
  },
  tagline: {
    color: '#486654',
    fontSize: 16,
    lineHeight: 24,
    marginTop: 12,
    maxWidth: 340,
  },
  form: {
    gap: 16,
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
  forgotButton: {
    alignSelf: 'flex-end',
    paddingVertical: 4,
  },
  forgotText: {
    color: '#1f7a4a',
    fontSize: 14,
    fontWeight: '700',
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
    minHeight: 54,
    justifyContent: 'center',
    marginTop: 8,
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
