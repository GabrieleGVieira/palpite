const features = [
  'Ranking dos Palpiteiros',
  'Bolões privados',
  'Palpites por partida',
  'Palpitômetro preparado',
  'Palpicoins em preparação',
  'Sugestões da PalpitAI',
  'Atualizações em tempo real',
];

export default function Features() {
  return (
    <section className="content-section alt-section" id="diferenciais" aria-labelledby="features-title">
      <div className="section-shell feature-layout">
        <div className="section-heading compact">
          <p className="eyebrow">Diferenciais</p>
          <h2 id="features-title">Feito para competir com amigos na Copa do Mundo 2026</h2>
          <p>
            Crie grupos privados, registre palpites, acompanhe rankings e mantenha a resenha
            viva sem planilhas ou conferência manual.
          </p>
        </div>
        <div className="feature-list" aria-label="Diferenciais do Palpite!">
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
