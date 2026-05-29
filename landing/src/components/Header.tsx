export default function Header() {
  return (
    <header className="site-header">
      <a className="brand" href="#top" aria-label="PalpitAI inicio">
        <img className="brand-mark" src="/splash-palpitai.png" alt="" width="34" height="34" />
        <span>PalpitAI</span>
      </a>
      <nav className="header-nav" aria-label="Navegação principal">
        <a href="#como-funciona">Como funciona</a>
        <a href="#diferenciais">Diferenciais</a>
        <a href="#faq">FAQ</a>
      </nav>
      <a className="header-cta" href="#teste-android">
        Entrar no Beta
      </a>
    </header>
  );
}
