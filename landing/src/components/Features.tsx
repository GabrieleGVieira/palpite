const features = [
  'Rankings automáticos',
  'Bolões privados',
  'Competições internacionais',
  'Previsões com IA',
  'Atualizações em tempo real',
  'Experiência simples para amigos e grupos',
];

export default function Features() {
  return (
    <section className="content-section alt-section" id="diferenciais" aria-labelledby="features-title">
      <div className="section-shell feature-layout">
        <div className="section-heading compact">
          <p className="eyebrow">Diferenciais</p>
          <h2 id="features-title">Feito para bolão online e para Copa do Mundo 2026</h2>
          <p>
            Combine grupos privados, app de palpites e previsões com IA em uma experiência simples
            para disputar com amigos sem planilhas ou conferência manual.
          </p>
        </div>
        <div className="feature-list" aria-label="Diferenciais do PalpitAI">
          {features.map((feature) => (
            <div className="feature-item" key={feature}>
              <span aria-hidden="true" />
              <p>{feature}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
