import LegalNotice from './LegalNotice';
import { appPath, assetPath } from '../assets';

const playStoreUrl = 'https://play.google.com/store/apps/details?id=com.palpitai.app';

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
            <a className="button button-primary" href={appPath('#teste-android')}>
              Baixar no Android
            </a>
            <a className="button button-secondary" href="https://palpite.expo.app" rel="noreferrer" target="_blank">
              Usar no iPhone
            </a>
          </div>
          <LegalNotice />
          <p className="button-note">No iPhone, abra no Safari, toque em Compartilhar e depois em Adicionar à Tela de Início.</p>
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
