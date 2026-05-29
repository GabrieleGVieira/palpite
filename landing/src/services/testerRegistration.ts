export type TesterRegistrationPayload = {
  name: string;
  email: string;
};

const storageKey = 'palpitai_android_testers';

export async function registerTester(payload: TesterRegistrationPayload): Promise<void> {
  const currentRegistrations = readRegistrations();
  const nextRegistrations = [
    ...currentRegistrations.filter((registration) => registration.email !== payload.email),
    {
      ...payload,
      registeredAt: new Date().toISOString(),
    },
  ];

  localStorage.setItem(storageKey, JSON.stringify(nextRegistrations));

  // Future integration point:
  // - send tester data to Supabase;
  // - sync approved Google emails with Google Groups;
  // - enroll or invite testers to Play Store Closed Testing.
  await Promise.resolve();
}

function readRegistrations(): Array<TesterRegistrationPayload & { registeredAt: string }> {
  const storedValue = localStorage.getItem(storageKey);

  if (!storedValue) {
    return [];
  }

  try {
    const parsedValue: unknown = JSON.parse(storedValue);

    if (!Array.isArray(parsedValue)) {
      return [];
    }

    return parsedValue.filter(isStoredRegistration);
  } catch {
    return [];
  }
}

function isStoredRegistration(
  value: unknown,
): value is TesterRegistrationPayload & { registeredAt: string } {
  if (!value || typeof value !== 'object') {
    return false;
  }

  const registration = value as Record<string, unknown>;

  return (
    typeof registration.name === 'string' &&
    typeof registration.email === 'string' &&
    typeof registration.registeredAt === 'string'
  );
}
