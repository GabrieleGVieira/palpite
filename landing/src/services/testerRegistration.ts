export type TesterRegistrationPayload = {
  name?: string;
  email: string;
  consent: boolean;
};

export type TesterRegistrationResult = {
  success: boolean;
  redirectUrl?: string;
  status?: string;
  message?: string;
};

const apiUrl = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '');

export async function registerTester(
  payload: TesterRegistrationPayload,
): Promise<TesterRegistrationResult> {
  const response = await fetch(`${apiUrl}/api/beta/android`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  let body: TesterRegistrationResult & { error?: string };
  try {
    body = await response.json();
  } catch {
    body = {
      success: false,
      message: 'Não foi possível processar a resposta do servidor.',
    };
  }

  if (!response.ok && response.status !== 202) {
    throw new Error(body.error ?? body.message ?? 'Não foi possível cadastrar seu e-mail.');
  }

  return body;
}
