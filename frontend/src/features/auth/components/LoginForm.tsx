import { Pressable, View, Text, StyleSheet } from 'react-native';
import { AuthInputField } from '../../../shared/components/AuthInputField';
import { FinishButton } from '../../../shared/components/FinishButton';
type LoginFormProps = {
  setEmail: (email: string) => void;
  setPassword: (password: string) => void;
  handleLogin: () => void;
  handleRecoverPassword: () => void;
  isSubmitting: boolean;
  formError: string | null;
  error: string | null;
  successMessage: string | null;
  isConfigured: boolean;
  email: string;
  password: string;
};

export function LoginForm({
  setEmail,
  email,
  setPassword,
  password,
  handleLogin,
  handleRecoverPassword,
  isSubmitting,
  formError,
  error,
  successMessage,
  isConfigured,
}: LoginFormProps) {
  return (
    <View style={styles.form}>
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
        placeholder="Sua senha"
        secureTextEntry
        value={password}
      />

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

      <FinishButton
        isLoading={isSubmitting || !isConfigured}
        onPress={handleLogin}
        loadingLabel="Entrando..."
        waitingLabel="Entrar"
      />
    </View>
  );
}

const styles = StyleSheet.create({
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
  form: {
    gap: 16,
    paddingTop: 24,
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
});
