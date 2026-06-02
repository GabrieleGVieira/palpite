import { FormEvent, useState } from 'react';
import { registerTester } from '../services/testerRegistration';

type FormErrors = {
  email?: string;
  consent?: string;
};

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const receivedMessage =
  'Cadastro recebido!\nVocê receberá acesso à versão beta do Palpite! assim que seu e-mail for aprovado na lista de testes.';

export default function TesterForm() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [consent, setConsent] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors: FormErrors = {};

    if (!email.trim()) {
      nextErrors.email = 'Informe seu e-mail Google.';
    } else if (!emailPattern.test(email.trim())) {
      nextErrors.email = 'Informe um e-mail válido.';
    }

    if (!consent) {
      nextErrors.consent = 'Confirme o consentimento para receber acesso beta e comunicações.';
    }

    setErrors(nextErrors);
    setSuccessMessage('');
    setErrorMessage('');

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await registerTester({
        name: name.trim() || undefined,
        email: email.trim().toLowerCase(),
        consent,
      });

      if (!result.success) {
        setErrorMessage(
          result.message ??
            'Recebemos seu e-mail, mas tivemos um problema ao liberar o acesso automaticamente. Tente novamente mais tarde.',
        );
        return;
      }

      setSuccessMessage(
        receivedMessage,
      );
      setName('');
      setEmail('');
      setConsent(false);
    } catch (error) {
      setErrorMessage(
        error instanceof Error
          ? error.message
          : 'Não foi possível cadastrar seu e-mail agora. Tente novamente mais tarde.',
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form className="tester-form" onSubmit={handleSubmit} noValidate>
      <h3>Entrar na lista Android</h3>
      <p>
        Use o e-mail Google que você pretende usar na Play Store para participar do Beta fechado.
      </p>

      <div className="form-field">
        <label htmlFor="tester-name">Nome opcional</label>
        <input
          id="tester-name"
          name="name"
          type="text"
          autoComplete="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
      </div>

      <div className="form-field">
        <label htmlFor="tester-email">E-mail Google</label>
        <input
          id="tester-email"
          name="email"
          type="email"
          inputMode="email"
          autoComplete="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          aria-invalid={Boolean(errors.email)}
          aria-describedby={errors.email ? 'tester-email-error' : undefined}
        />
        {errors.email ? (
          <span className="field-error" id="tester-email-error">
            {errors.email}
          </span>
        ) : null}
      </div>

      <label className="form-checkbox" htmlFor="tester-consent">
        <input
          id="tester-consent"
          name="consent"
          type="checkbox"
          checked={consent}
          onChange={(event) => setConsent(event.target.checked)}
          aria-invalid={Boolean(errors.consent)}
          aria-describedby={errors.consent ? 'tester-consent-error' : undefined}
        />
        <span>Autorizo receber acesso beta e comunicações sobre o Palpite!</span>
      </label>
      {errors.consent ? (
        <span className="field-error" id="tester-consent-error">
          {errors.consent}
        </span>
      ) : null}

      <button className="button button-primary form-submit" type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Enviando...' : successMessage ? 'Cadastro recebido' : 'Cadastrar e-mail Google'}
      </button>

      {successMessage ? (
        <p className="success-message" role="status">
          {successMessage}
        </p>
      ) : null}

      {errorMessage ? (
        <p className="error-message" role="alert">
          {errorMessage}
        </p>
      ) : null}
    </form>
  );
}
