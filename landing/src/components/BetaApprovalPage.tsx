import { useEffect, useMemo, useState } from 'react';
import {
  BetaApprovalPreview,
  confirmBetaApproval,
  getBetaApprovalPreview,
} from '../services/betaApproval';

type BetaApprovalPageProps = {
  testerId: string;
};

type ViewState = 'loading' | 'ready' | 'submitting' | 'success' | 'error';

export default function BetaApprovalPage({ testerId }: BetaApprovalPageProps) {
  const token = useMemo(() => new URLSearchParams(window.location.search).get('token') ?? '', []);
  const [preview, setPreview] = useState<BetaApprovalPreview | null>(null);
  const [status, setStatus] = useState<ViewState>('loading');
  const [message, setMessage] = useState('');

  useEffect(() => {
    let isMounted = true;

    async function loadPreview() {
      if (!testerId || !token) {
        setStatus('error');
        setMessage('Link de aprovação inválido.');
        return;
      }

      try {
        const data = await getBetaApprovalPreview(testerId, token);
        if (!isMounted) {
          return;
        }
        setPreview(data);
        setStatus('ready');
      } catch (error) {
        if (!isMounted) {
          return;
        }
        setStatus('error');
        setMessage(error instanceof Error ? error.message : 'Não foi possível carregar a aprovação.');
      }
    }

    loadPreview();
    return () => {
      isMounted = false;
    };
  }, [testerId, token]);

  async function handleConfirm() {
    if (!preview) {
      return;
    }

    setStatus('submitting');
    setMessage('');

    try {
      const result = await confirmBetaApproval(preview.testerId, token);
      setStatus('success');
      setMessage(`Aprovação confirmada. Status: ${result.status}.`);
    } catch (error) {
      setStatus('error');
      setMessage(error instanceof Error ? error.message : 'Não foi possível confirmar a aprovação.');
    }
  }

  return (
    <main className="approval-page">
      <section className="section-shell approval-panel" aria-labelledby="approval-title">
        <p className="eyebrow">Beta Android</p>
        <h1 id="approval-title">Confirmar aprovação</h1>

        {status === 'loading' ? <p className="approval-muted">Carregando cadastro...</p> : null}

        {preview ? (
          <div className="approval-summary" aria-live="polite">
            <p>Você está prestes a aprovar:</p>
            <dl>
              <div>
                <dt>Nome</dt>
                <dd>{preview.name || 'Não informado'}</dd>
              </div>
              <div>
                <dt>Email</dt>
                <dd>{preview.email}</dd>
              </div>
              <div>
                <dt>Status atual</dt>
                <dd>{preview.status}</dd>
              </div>
            </dl>
            <p className="approval-muted">
              Confirme que este e-mail já foi adicionado no Play Console.
            </p>
          </div>
        ) : null}

        {status === 'ready' || status === 'submitting' ? (
          <button
            className="button button-primary approval-button"
            type="button"
            disabled={status === 'submitting'}
            onClick={handleConfirm}
          >
            {status === 'submitting' ? 'Confirmando...' : 'Confirmar aprovação'}
          </button>
        ) : null}

        {status === 'success' ? (
          <p className="success-message" role="status">
            {message}
          </p>
        ) : null}

        {status === 'error' ? (
          <p className="error-message" role="alert">
            {message}
          </p>
        ) : null}
      </section>
    </main>
  );
}
