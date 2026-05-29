import { appPath } from '../assets';

export default function LegalNotice({ className = '' }: { className?: string }) {
  return (
    <p className={`legal-notice ${className}`.trim()}>
      Ao continuar, você concorda com os{' '}
      <a href={appPath('terms')}>Termos de Uso</a> e a{' '}
      <a href={appPath('privacy')}>Política de Privacidade</a>.
    </p>
  );
}
