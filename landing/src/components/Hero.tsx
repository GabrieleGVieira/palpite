import LegalNotice from './LegalNotice';
import { assetPath } from '../assets';

export default function Hero() {
  return (
    <section className="hero-section" id="top" aria-labelledby="hero-title">
      <div className="section-shell hero-grid">
        <div className="hero-copy">
          <p className="eyebrow">Bolão, ranking e resenha</p>
          <h1 id="hero-title">Palpite!</h1>
          <p className="hero-subtitle">Onde a resenha vira competição.</p>
          <p className="hero-text">
            Mostre quem entende de futebol, desafie seus amigos e acompanhe rankings que deixam
            cada jogo mais divertido.
          </p>
          <div className="hero-actions" aria-label="Ações da landing">
            <a className="button button-primary" href="#teste-android">
              Entrar no Beta Android
            </a>
            <button className="button button-disabled" type="button" disabled>
              iPhone em breve
            </button>
          </div>
          <LegalNotice />
          <p className="button-note">Versão para iPhone será liberada em breve.</p>
          <p className="iphone-note">Estamos preparando a melhor experiência para iPhone.</p>
        </div>
        <div className="hero-visual" aria-label="Visual do Palpite!">
          <img
            className="hero-illustration"
            src={assetPath('/landing-hero.png')}
            alt=""
            width="1200"
            height="630"
          />
          <img
            className="hero-logo-badge"
            src={assetPath('/splash-palpite.png')}
            alt=""
            width="700"
            height="700"
          />
        </div>
      </div>
    </section>
  );
}
