export default function Hero() {
  return (
    <section className="hero-section" id="top" aria-labelledby="hero-title">
      <div className="section-shell hero-grid">
        <div className="hero-copy">
          <p className="eyebrow">Bolão de futebol inteligente</p>
          <h1 id="hero-title">PalpitAI</h1>
          <p className="hero-subtitle">App de bolão com IA para futebol</p>
          <p className="hero-text">
            Crie bolões, dispute com amigos e receba previsões inteligentes para ajudar nos seus
            palpites.
          </p>
          <div className="hero-actions" aria-label="Ações da landing">
            <a className="button button-primary" href="#teste-android">
              Entrar no Beta Android
            </a>
            <button className="button button-disabled" type="button" disabled>
              iPhone em breve
            </button>
          </div>
          <p className="button-note">Versão para iPhone será liberada em breve.</p>
          <p className="iphone-note">Estamos preparando a melhor experiência para iPhone.</p>
        </div>
        <div className="hero-visual" aria-label="Visual do PalpitAI">
          <img className="hero-illustration" src="/landing-hero.png" alt="" width="1200" height="630" />
          <img className="hero-logo-badge" src="/splash-palpitai.png" alt="" width="700" height="700" />
        </div>
      </div>
    </section>
  );
}
