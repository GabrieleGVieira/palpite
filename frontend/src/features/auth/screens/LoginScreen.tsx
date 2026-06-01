import { StatusBar } from 'expo-status-bar';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useLoginScreen } from '../hooks/useLoginScreen';
import { Footer } from '../../../shared/components/Footer';
import { Header } from '../../../shared/components/Header';
import { LegalConsentText } from '../../../shared/components/LegalConsentText';
import { LoginForm } from '../components/LoginForm';

type LoginScreenProps = {
  onCreateAccount: () => void;
};

export function LoginScreen({ onCreateAccount }: LoginScreenProps) {
  const {
    email,
    error,
    formError,
    handleLogin,
    handleRecoverPassword,
    isConfigured,
    isSubmitting,
    password,
    setEmail,
    setPassword,
    successMessage,
  } = useLoginScreen();

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="dark" />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.keyboardView}>
        <ScrollView
          contentContainerStyle={styles.container}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}>
          <View style={styles.pitchMarkTop} />
          <View style={styles.pitchCircle} />

          <Header
            title="Bem-vindo ao Palpite!"
            subtitle="Entre no seu bolão inteligente e acompanhe palpites, rankings e insights com IA."
            topSpacing={32}
          />

          <LoginForm
            setEmail={setEmail}
            setPassword={setPassword}
            handleLogin={handleLogin}
            handleRecoverPassword={handleRecoverPassword}
            isSubmitting={isSubmitting}
            formError={formError}
            error={error}
            successMessage={successMessage}
            isConfigured={isConfigured}
            email={email}
            password={password}
          />

          <LegalConsentText />

          <Footer
            question="Ainda não tem conta?"
            buttonLabel="Criar conta"
            onButtonPress={onCreateAccount}
          />
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
  container: {
    flexGrow: 1,
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
});
