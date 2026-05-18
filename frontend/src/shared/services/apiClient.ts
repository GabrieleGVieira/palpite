import { supabase } from '../../services/supabase';

export const apiURL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:3000';
const defaultTimeoutMs = 15_000;

type APIError = {
  error?: string;
};

type RequestOptions = RequestInit & {
  fallbackError: string;
};

export async function apiClient<T>(path: string, options: RequestOptions): Promise<T> {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new Error('Sua sessao expirou. Entre novamente.');
  }

  let response: Response;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), defaultTimeoutMs);

  try {
    response = await fetch(`${apiURL}${path}`, {
      ...options,
      signal: controller.signal,
      headers: {
        Authorization: `Bearer ${session.access_token}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });
  } catch {
    throw new Error(
      `Não foi possivel acessar a API em ${apiURL}. Verifique se o backend esta rodando e se o celular esta na mesma rede.`,
    );
  } finally {
    clearTimeout(timeout);
  }

  const data = await parseJSON<T | APIError>(response);

  if (!response.ok) {
    const apiError = data as APIError;
    throw new Error(apiError.error ? apiError.error : options.fallbackError);
  }

  return data as T;
}

async function parseJSON<T>(response: Response): Promise<T> {
  const responseText = await response.text();

  if (!responseText) {
    return {} as T;
  }

  try {
    return JSON.parse(responseText) as T;
  } catch {
    throw new Error(`A API respondeu em formato inesperado: ${responseText}`);
  }
}
