import { assetPath } from '../assets';

const steps = [
  {
    title: 'Crie grupos',
    text: 'Monte bolões privados para seus amigos.',
    image: '/onboarding-groups.png',
  },
  {
    title: 'Dê seu palpite',
    text: 'Registre placares nas partidas e competições.',
    image: '/make-predictions.png',
  },
  {
    title: 'Suba no ranking',
    text: 'Acompanhe a disputa dos Palpiteiros.',
    image: '/onboarding-live.png',
  },
  {
    title: 'Consulte a PalpitAI',
    text: 'Veja análises e tendências como apoio complementar.',
    image: '/onboarding-ai.png',
  },
];

export default function HowItWorks() {
  return (
    <section className="content-section" id="como-funciona" aria-labelledby="how-title">
      <div className="section-shell">
        <div className="section-heading">
          <p className="eyebrow">Como funciona</p>
          <h2 id="how-title">Seu bolão com amigos em poucos passos</h2>
          <p>
            O Palpite! organiza o bolão de futebol, centraliza os palpites e mantém o ranking
            atualizado para todo mundo acompanhar.
          </p>
        </div>
        <div className="card-grid four-columns">
          {steps.map((step) => (
            <article className="info-card" key={step.title}>
              <div className="phone-preview" aria-hidden="true">
                <span className="phone-speaker" />
                <img src={assetPath(step.image)} alt="" width="600" height="900" />
              </div>
              <h3>{step.title}</h3>
              <p>{step.text}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
