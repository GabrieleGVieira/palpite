import { StyleSheet, Text, View } from 'react-native';
import { AuthInputField } from '../../../shared/components/AuthInputField';
import { FinishButton } from '../../../shared/components/FinishButton';

type SignupFormProps = {
  setName: (name: string) => void;
  name: string;
  setEmail: (email: string) => void;
  email: string;
  setPassword: (password: string) => void;
  password: string;
  setConfirmPassword: (confirmPassword: string) => void;
  confirmPassword: string;
  handleSignup: (onBackToLogin: () => void) => Promise<void>;
  isSubmitting: boolean;
  formError: string | null;
  error: string | null;
  isConfigured: boolean;
  onBackToLogin: () => void;
};

export function SignupForm({
  setName,
  name,
  setEmail,
  email,
  setPassword,
  password,
  setConfirmPassword,
  confirmPassword,
  handleSignup,
  isSubmitting,
  formError,
  error,
  isConfigured,
  onBackToLogin,
}: SignupFormProps) {
  return (
    <View style={styles.form}>
      <AuthInputField
        autoCapitalize="words"
        keyboardType="default"
        label="Nome"
        onChangeText={setName}
        placeholder="Seu nome"
        value={name}
      />

      <AuthInputField
        autoCapitalize="none"
        keyboardType="email-address"
        label="E-mail"
        onChangeText={setEmail}
        placeholder="voce@email.com"
        value={email}
      />

      <AuthInputField
        autoCapitalize="none"
        keyboardType="default"
        label="Senha"
        onChangeText={setPassword}
        placeholder="Crie uma senha"
        secureTextEntry
        value={password}
      />

      <AuthInputField
        autoCapitalize="none"
        keyboardType="default"
        label="Confirmar senha"
        onChangeText={setConfirmPassword}
        placeholder="Repita sua senha"
        secureTextEntry
        value={confirmPassword}
      />

      {!isConfigured ? (
        <Text style={styles.feedbackText}>
          Configure EXPO_PUBLIC_SUPABASE_URL e EXPO_PUBLIC_SUPABASE_KEY.
        </Text>
      ) : null}

      {formError ? <Text style={styles.errorText}>{formError}</Text> : null}
      {error ? <Text style={styles.errorText}>{error}</Text> : null}

      <FinishButton
        isLoading={isSubmitting || !isConfigured}
        onPress={() => handleSignup(onBackToLogin)}
        loadingLabel="Criando conta..."
        waitingLabel="Criar conta"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  form: {
    gap: 16,
    marginTop: 40,
  },
  feedbackText: {
    color: '#8a5f00',
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
  errorText: {
    color: '#a03222',
    fontSize: 13,
    lineHeight: 18,
  },
});
