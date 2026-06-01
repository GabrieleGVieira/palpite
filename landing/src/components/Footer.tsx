import { appPath } from '../assets';

export default function Footer() {
  return (
    <footer className="site-footer">
      <div className="section-shell footer-content">
        <div>
          <p>Palpite!</p>
          <p>Bolões, rankings e resenha para quem vive futebol. Beta Android, iPhone em breve.</p>
        </div>
        <nav className="footer-links" aria-label="Links legais">
          <a href={appPath('privacy')}>Política de Privacidade</a>
          <a href={appPath('terms')}>Termos de Uso</a>
          <a href={appPath('account-deletion')}>Exclusão de Conta</a>
        </nav>
      </div>
    </footer>
  );
}
