import { useState } from 'react';
import { useAuth } from './useAuth';

export function useSignupScreen() {
  const { clearError, error, isConfigured, isSubmitting, signup } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function handleSignup(onSuccess: () => void) {
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
      onSuccess();
    }
  }

  return {
    name,
    setName,
    email,
    setEmail,
    password,
    setPassword,
    confirmPassword,
    setConfirmPassword,
    formError,
    successMessage,
    error,
    isConfigured,
    isSubmitting,
    handleSignup,
  };
}
