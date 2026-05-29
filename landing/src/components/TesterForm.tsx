import { FormEvent, useState } from 'react';
import { registerTester } from '../services/testerRegistration';

type FormErrors = {
  name?: string;
  email?: string;
};

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function TesterForm() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [errors, setErrors] = useState<FormErrors>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors: FormErrors = {};

    if (!name.trim()) {
      nextErrors.name = 'Informe seu nome.';
    }

    if (!email.trim()) {
      nextErrors.email = 'Informe seu e-mail Google.';
    } else if (!emailPattern.test(email.trim())) {
      nextErrors.email = 'Informe um e-mail válido.';
    }

    setErrors(nextErrors);
    setSuccessMessage('');

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setIsSubmitting(true);
    await registerTester({ name: name.trim(), email: email.trim().toLowerCase() });
    setIsSubmitting(false);
    setName('');
    setEmail('');
    setSuccessMessage('Cadastro realizado! Em breve você receberá acesso ao Beta.');
  }

  return (
    <form className="tester-form" onSubmit={handleSubmit} noValidate>
      <h3>Entrar na lista Android</h3>
      <p>
        Use o e-mail Google que você pretende usar na Play Store para participar do Beta fechado.
      </p>

      <div className="form-field">
        <label htmlFor="tester-name">Nome</label>
        <input
          id="tester-name"
          name="name"
          type="text"
          autoComplete="name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          aria-invalid={Boolean(errors.name)}
          aria-describedby={errors.name ? 'tester-name-error' : undefined}
        />
        {errors.name ? (
          <span className="field-error" id="tester-name-error">
            {errors.name}
          </span>
        ) : null}
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

      <button className="button button-primary form-submit" type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Enviando...' : 'Cadastrar e-mail Google'}
      </button>

      {successMessage ? (
        <p className="success-message" role="status">
          {successMessage}
        </p>
      ) : null}
    </form>
  );
}
