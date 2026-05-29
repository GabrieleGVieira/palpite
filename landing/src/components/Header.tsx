import { appPath, assetPath } from '../assets';

export default function Header() {
  const basePath = appPath();
  const isHome = window.location.pathname === basePath || window.location.pathname === basePath.replace(/\/$/, '');
  const homeHref = isHome ? '#top' : basePath;

  return (
    <header className="site-header">
      <a className="brand" href={homeHref} aria-label="PalpitAI inicio">
        <img
          className="brand-mark"
          src={assetPath('/splash-palpitai.png')}
          alt=""
          width="34"
          height="34"
        />
        <span>PalpitAI</span>
      </a>
      <nav className="header-nav" aria-label="Navegação principal">
        <a href={appPath('#como-funciona')}>Como funciona</a>
        <a href={appPath('#diferenciais')}>Diferenciais</a>
        <a href={appPath('#faq')}>FAQ</a>
      </nav>
      <a className="header-cta" href={appPath('#teste-android')}>
        Entrar no Beta
      </a>
    </header>
  );
}
