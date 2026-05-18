import { useState } from 'react';
import { useAuth } from './useAuth';

export function useLoginScreen() {
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

  return {
    email,
    setEmail,
    password,
    setPassword,
    formError,
    successMessage,
    error,
    isConfigured,
    isSubmitting,
    handleLogin,
    handleRecoverPassword,
  };
}
