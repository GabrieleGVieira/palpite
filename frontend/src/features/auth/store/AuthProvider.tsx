import type { Session, User } from '@supabase/supabase-js';
import { useQueryClient } from '@tanstack/react-query';
import { createContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react';

import {
  getAuthErrorMessage,
  sendPasswordResetEmail,
  signInWithEmail,
  signOut,
  signUpWithEmail,
} from '../services/auth';
import { isSupabaseConfigured, supabase } from '../../../services/supabase';

type AuthContextValue = {
  error: string | null;
  isConfigured: boolean;
  isLoading: boolean;
  isSubmitting: boolean;
  session: Session | null;
  user: User | null;
  clearError: () => void;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  recoverPassword: (email: string) => Promise<boolean>;
  signup: (name: string, email: string, password: string) => Promise<boolean>;
};

export const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient();
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function loadSession() {
      if (!isSupabaseConfigured) {
        setIsLoading(false);
        return;
      }

      const { data, error: sessionError } = await supabase.auth.getSession();

      if (!isMounted) {
        return;
      }

      if (sessionError) {
        setError(getAuthErrorMessage(sessionError));
      }

      setSession(data.session);
      setIsLoading(false);
    }

    loadSession();

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, nextSession) => {
      if (!nextSession) {
        queryClient.clear();
      }
      setSession(nextSession);
    });

    return () => {
      isMounted = false;
      subscription.unsubscribe();
    };
  }, [queryClient]);

  const value = useMemo<AuthContextValue>(
    () => ({
      clearError: () => setError(null),
      error,
      isConfigured: isSupabaseConfigured,
      isLoading,
      isSubmitting,
      login: async (email, password) => {
        setError(null);
        setIsSubmitting(true);

        try {
          await signInWithEmail({ email, password });
          return true;
        } catch (authError) {
          setError(getAuthErrorMessage(authError));
          return false;
        } finally {
          setIsSubmitting(false);
        }
      },
      logout: async () => {
        setError(null);
        setIsSubmitting(true);

        try {
          await signOut();
          queryClient.clear();
        } catch (authError) {
          setError(getAuthErrorMessage(authError));
        } finally {
          setIsSubmitting(false);
        }
      },
      recoverPassword: async (email) => {
        setError(null);
        setIsSubmitting(true);

        try {
          await sendPasswordResetEmail(email);
          return true;
        } catch (authError) {
          setError(getAuthErrorMessage(authError));
          return false;
        } finally {
          setIsSubmitting(false);
        }
      },
      session,
      signup: async (name, email, password) => {
        setError(null);
        setIsSubmitting(true);

        try {
          await signUpWithEmail({ email, name, password });
          return true;
        } catch (authError) {
          setError(getAuthErrorMessage(authError));
          return false;
        } finally {
          setIsSubmitting(false);
        }
      },
      user: session?.user ?? null,
    }),
    [error, isLoading, isSubmitting, queryClient, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
