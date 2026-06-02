export type BetaApprovalPreview = {
  testerId: string;
  name: string;
  email: string;
  status: string;
};

export type BetaApprovalResult = {
  success: boolean;
  status: string;
};

const apiUrl = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '');

async function readJSON<T>(response: Response): Promise<T & { error?: string }> {
  try {
    return await response.json();
  } catch {
    return {
      error: 'Não foi possível processar a resposta do servidor.',
    } as T & { error?: string };
  }
}

export async function getBetaApprovalPreview(
  testerId: string,
  token: string,
): Promise<BetaApprovalPreview> {
  const response = await fetch(
    `${apiUrl}/api/admin/beta-testers/${encodeURIComponent(testerId)}/approve/preview?token=${encodeURIComponent(token)}`,
  );
  const body = await readJSON<BetaApprovalPreview>(response);

  if (!response.ok) {
    throw new Error(body.error ?? 'Não foi possível carregar a aprovação.');
  }

  return body;
}

export async function confirmBetaApproval(
  testerId: string,
  token: string,
): Promise<BetaApprovalResult> {
  const response = await fetch(
    `${apiUrl}/api/admin/beta-testers/${encodeURIComponent(testerId)}/approve/confirm?token=${encodeURIComponent(token)}`,
    {
      method: 'POST',
    },
  );
  const body = await readJSON<BetaApprovalResult>(response);

  if (!response.ok) {
    throw new Error(body.error ?? 'Não foi possível confirmar a aprovação.');
  }

  return body;
}
