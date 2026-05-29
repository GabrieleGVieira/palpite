const steps = [
  {
    title: 'Crie grupos',
    text: 'Monte bolões privados para seus amigos.',
    image: '/onboarding-groups.png',
  },
  {
    title: 'Faça palpites',
    text: 'Palpite em partidas e competições.',
    image: '/onboarding-live.png',
  },
  {
    title: 'Ganhe pontos',
    text: 'Acompanhe rankings automáticos.',
    image: '/onboarding-live.png',
  },
  {
    title: 'Use IA',
    text: 'Receba análises e previsões para ajudar suas decisões.',
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
            O PalpitAI organiza o bolão de futebol, centraliza os palpites de futebol e mantém o
            ranking de palpites atualizado para todo mundo acompanhar.
          </p>
        </div>
        <div className="card-grid four-columns">
          {steps.map((step) => (
            <article className="info-card" key={step.title}>
              <div className="phone-preview" aria-hidden="true">
                <span className="phone-speaker" />
                <img src={step.image} alt="" width="600" height="900" />
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
