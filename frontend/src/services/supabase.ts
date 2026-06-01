import AsyncStorage from '@react-native-async-storage/async-storage';
import { createClient, processLock } from '@supabase/supabase-js';
import { AppState } from 'react-native';
import 'react-native-url-polyfill/auto';

const supabaseUrl = process.env.EXPO_PUBLIC_SUPABASE_URL;
const supabaseKey =
  process.env.EXPO_PUBLIC_SUPABASE_KEY ?? process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY;
const isStaticRendering = typeof window === 'undefined';

export const isSupabaseConfigured = Boolean(supabaseUrl && supabaseKey);

export const supabase = createClient(
  supabaseUrl ?? 'https://example.supabase.co',
  supabaseKey ?? 'missing-supabase-key',
  {
    auth: {
      autoRefreshToken: !isStaticRendering,
      detectSessionInUrl: false,
      lock: processLock,
      persistSession: !isStaticRendering,
      storage: isStaticRendering ? undefined : AsyncStorage,
    },
  },
);

if (!isStaticRendering) {
  AppState.addEventListener('change', (state) => {
    if (state === 'active') {
      supabase.auth.startAutoRefresh();
      return;
    }

    supabase.auth.stopAutoRefresh();
  });
}
