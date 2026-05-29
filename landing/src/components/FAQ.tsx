const questions = [
  {
    question: 'O que é o PalpitAI?',
    answer:
      'PalpitAI é um app de bolão de futebol com IA para criar grupos, registrar palpites e acompanhar rankings.',
  },
  {
    question: 'Como funciona o app de bolão?',
    answer:
      'Você cria ou entra em um bolão online, faz palpites de futebol nas partidas e soma pontos automaticamente no ranking.',
  },
  {
    question: 'O PalpitAI usa inteligência artificial?',
    answer:
      'Sim. O app usa previsões com IA para apoiar suas decisões antes de enviar os palpites.',
  },
  {
    question: 'Posso criar bolões com amigos?',
    answer:
      'Sim. O PalpitAI foi pensado para bolão com amigos, grupos privados e rankings simples de acompanhar.',
  },
  {
    question: 'O app está disponível para iPhone?',
    answer:
      'A versão para iPhone será liberada em breve. No momento, o foco do Beta antecipado é Android.',
  },
  {
    question: 'Como participar do Beta no Android?',
    answer:
      'Cadastre seu nome e e-mail Google na landing page para receber acesso ao Beta Android quando a lista for liberada.',
  },
];

export default function FAQ() {
  return (
    <section className="content-section faq-section" id="faq" aria-labelledby="faq-title">
      <div className="section-shell">
        <div className="section-heading">
          <p className="eyebrow">FAQ</p>
          <h2 id="faq-title">Perguntas frequentes</h2>
        </div>
        <div className="faq-list">
          {questions.map((item) => (
            <article className="faq-item" key={item.question}>
              <h3>{item.question}</h3>
              <p>{item.answer}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
