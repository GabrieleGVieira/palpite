import type { AuthError } from '@supabase/supabase-js';

import { isSupabaseConfigured, supabase } from './supabase';

type SignInPayload = {
  email: string;
  password: string;
};

type SignUpPayload = SignInPayload & {
  name: string;
};

function ensureSupabaseConfigured() {
  if (!isSupabaseConfigured) {
    throw new Error(
      'Configure EXPO_PUBLIC_SUPABASE_URL e EXPO_PUBLIC_SUPABASE_KEY para usar autenticacao.',
    );
  }
}

export function getAuthErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Não foi possivel concluir a acao. Tente novamente.';
}

export async function signInWithEmail({ email, password }: SignInPayload) {
  ensureSupabaseConfigured();

  const { error } = await supabase.auth.signInWithPassword({
    email: email.trim(),
    password,
  });

  if (error) {
    throw error;
  }
}

export async function signUpWithEmail({ email, name, password }: SignUpPayload) {
  ensureSupabaseConfigured();

  const { data, error } = await supabase.auth.signUp({
    email: email.trim(),
    options: {
      data: {
        full_name: name.trim(),
      },
    },
    password,
  });

  if (error) {
    throw error;
  }

  return data;
}

export async function signOut() {
  ensureSupabaseConfigured();

  const { error } = await supabase.auth.signOut();

  if (error) {
    throw error;
  }
}

export async function sendPasswordResetEmail(email: string) {
  ensureSupabaseConfigured();

  const { error } = await supabase.auth.resetPasswordForEmail(email.trim());

  if (error) {
    throw error;
  }
}

export function isInvalidCredentialsError(error: unknown) {
  return (error as AuthError | undefined)?.message?.toLowerCase().includes('invalid login');
}
