import { StatusBar } from 'expo-status-bar';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useSignupScreen } from '../hooks/useSignupScreen';
import { Footer } from '../../../shared/components/Footer';
import { Header } from '../../../shared/components/Header';
import { LegalConsentText } from '../../../shared/components/LegalConsentText';
import { SignupForm } from '../components/SignupForm';
import { BackButton } from '../../../shared/components/BackButton';

type SignupScreenProps = {
  onBackToLogin: () => void;
};

export function SignupScreen({ onBackToLogin }: SignupScreenProps) {
  const {
    confirmPassword,
    email,
    error,
    formError,
    isConfigured,
    isSubmitting,
    name,
    password,
    setConfirmPassword,
    setEmail,
    setName,
    setPassword,
    handleSignup,
  } = useSignupScreen();

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

          <BackButton onPress={onBackToLogin} />

          <Header
            title="Crie sua conta"
            subtitle="Monte seus grupos, registre palpites e acompanhe os insights do Cravou! durante a competição."
            topSpacing={20}
          />

          <SignupForm
            confirmPassword={confirmPassword}
            email={email}
            error={error}
            formError={formError}
            isConfigured={isConfigured}
            isSubmitting={isSubmitting}
            name={name}
            password={password}
            setConfirmPassword={setConfirmPassword}
            setEmail={setEmail}
            setName={setName}
            setPassword={setPassword}
            handleSignup={handleSignup}
            onBackToLogin={onBackToLogin}
          />

          <LegalConsentText />

          <Footer
            question="Já tem uma conta?"
            buttonLabel="Faça login"
            onButtonPress={onBackToLogin}
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
  form: {
    gap: 16,
    marginTop: 40,
  },
});
