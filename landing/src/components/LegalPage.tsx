import { appPath } from '../assets';

type LegalSection = {
  title: string;
  paragraphs?: string[];
  bullets?: string[];
};

const privacySections: LegalSection[] = [
  {
    title: '1. Introdução',
    paragraphs: [
      'Bem-vindo ao PalpitAI.',
      'O PalpitAI é uma plataforma de bolões esportivos que permite a criação e participação em grupos de palpites, rankings e competições esportivas.',
      'Esta Política de Privacidade descreve como coletamos, utilizamos, armazenamos e protegemos suas informações.',
      'Ao utilizar o aplicativo, você concorda com os termos descritos neste documento.',
    ],
  },
  {
    title: '2. Dados coletados',
    paragraphs: [
      'Podemos coletar as seguintes informações:',
      'Dados de cadastro: nome, endereço de e-mail, foto de perfil e identificador único da conta.',
      'Dados de autenticação: quando o usuário realiza login por provedores externos, como Google, recebemos apenas as informações autorizadas pelo próprio usuário.',
      'Dados de uso: dispositivo utilizado, sistema operacional, versão do aplicativo, data e horário de acesso, eventos de navegação e informações relacionadas ao uso das funcionalidades do aplicativo.',
      'Dados relacionados aos bolões: palpites realizados, participação em grupos, rankings e histórico de resultados.',
      'Notificações: caso autorizado pelo usuário, poderemos armazenar identificadores de notificação para envio de avisos e atualizações.',
    ],
  },
  {
    title: '3. Como utilizamos seus dados',
    paragraphs: ['Os dados coletados são utilizados para:'],
    bullets: [
      'Permitir autenticação e acesso à conta',
      'Gerenciar grupos e bolões',
      'Calcular pontuações e rankings',
      'Exibir estatísticas e histórico',
      'Melhorar a experiência do usuário',
      'Corrigir falhas e problemas técnicos',
      'Enviar notificações relacionadas ao funcionamento do aplicativo',
      'Prevenir fraudes e uso indevido da plataforma',
    ],
  },
  {
    title: '4. Compartilhamento de informações',
    paragraphs: [
      'Não comercializamos dados pessoais.',
      'Os dados poderão ser processados por fornecedores responsáveis pela infraestrutura do serviço, incluindo serviços de autenticação, hospedagem, banco de dados, análise e monitoramento, e envio de notificações.',
      'Esses fornecedores possuem acesso apenas às informações necessárias para execução de suas funções.',
    ],
  },
  {
    title: '5. Segurança',
    paragraphs: [
      'Adotamos medidas técnicas e administrativas razoáveis para proteger os dados dos usuários contra acesso não autorizado, alteração indevida, divulgação indevida e destruição ou perda de informações.',
      'Embora empreguemos mecanismos de segurança, nenhum sistema é totalmente imune a riscos.',
    ],
  },
  {
    title: '6. Retenção de dados',
    paragraphs: [
      'Os dados serão mantidos enquanto a conta permanecer ativa ou enquanto forem necessários para o funcionamento do serviço e cumprimento de obrigações legais.',
    ],
  },
  {
    title: '7. Exclusão da conta',
    paragraphs: [
      'O usuário poderá solicitar a exclusão de sua conta e dos dados associados.',
      'A solicitação poderá ser realizada por meio dos canais de suporte disponibilizados pelo PalpitAI.',
      'Algumas informações poderão ser mantidas quando exigidas por obrigação legal ou regulatória.',
    ],
  },
  {
    title: '8. Menores de idade',
    paragraphs: [
      'O aplicativo não é destinado a menores de 13 anos.',
      'Caso seja identificado o tratamento indevido de dados de menores sem autorização adequada, as informações poderão ser removidas.',
    ],
  },
  {
    title: '9. Alterações nesta política',
    paragraphs: [
      'Esta Política de Privacidade poderá ser atualizada periodicamente.',
      'As alterações entrarão em vigor após sua publicação.',
    ],
  },
  {
    title: '10. Contato',
    paragraphs: [
      'Em caso de dúvidas sobre esta Política de Privacidade ou sobre o tratamento de dados pessoais, entre em contato pelo e-mail contato@palpitai.app.',
    ],
  },
];

const termsSections: LegalSection[] = [
  {
    title: '1. Aceitação dos termos',
    paragraphs: [
      'Ao acessar ou utilizar o PalpitAI, você declara que leu, compreendeu e concorda com estes Termos de Uso.',
      'Se você não concordar com estes termos, não deverá utilizar o aplicativo.',
    ],
  },
  {
    title: '2. Sobre o PalpitAI',
    paragraphs: [
      'O PalpitAI é uma plataforma de bolões esportivos que permite criar grupos, registrar palpites, acompanhar rankings, consultar estatísticas e visualizar previsões com IA.',
      'O PalpitAI não é uma plataforma oficial de apostas, jogos de azar ou transações financeiras.',
    ],
  },
  {
    title: '3. Cadastro e conta do usuário',
    paragraphs: [
      'Para utilizar determinadas funcionalidades, o usuário poderá precisar criar uma conta ou realizar login por provedores externos, como Google.',
      'O usuário é responsável por manter seus dados de acesso protegidos e por informar dados corretos no cadastro.',
    ],
  },
  {
    title: '4. Uso permitido da plataforma',
    paragraphs: [
      'O usuário deve utilizar o PalpitAI de forma lícita, respeitosa e compatível com a finalidade do aplicativo.',
      'É proibido tentar acessar áreas restritas, interferir no funcionamento da plataforma, explorar falhas técnicas ou utilizar o app para atividades ilegais.',
    ],
  },
  {
    title: '5. Bolões, palpites e rankings',
    paragraphs: [
      'Os palpites e rankings têm finalidade recreativa e competitiva entre usuários e grupos.',
      'Rankings, pontuações e critérios de desempate dependem das regras configuradas pelo app ou pelo grupo, podendo ser ajustados conforme evolução do serviço.',
      'O usuário é responsável pelos palpites enviados em sua conta.',
    ],
  },
  {
    title: '6. Previsões, estatísticas e IA',
    paragraphs: [
      'As previsões com IA, estatísticas e análises exibidas pelo PalpitAI são apenas informativas e não garantem resultados esportivos.',
      'Essas informações podem conter limitações, imprecisões ou variações conforme dados disponíveis, modelos utilizados e eventos externos.',
      'O usuário não deve tomar decisões financeiras, de apostas ou de risco com base exclusiva nas previsões do PalpitAI.',
    ],
  },
  {
    title: '7. Recursos pagos futuros',
    paragraphs: [
      'O PalpitAI poderá oferecer recursos pagos, assinaturas ou funcionalidades premium futuramente.',
      'Quando isso ocorrer, as condições comerciais, preços, formas de pagamento e regras específicas serão apresentadas antes da contratação.',
    ],
  },
  {
    title: '8. Responsabilidades do usuário',
    paragraphs: [
      'O usuário é responsável pelo conteúdo que envia, pelas interações realizadas em grupos e pelo uso adequado da plataforma.',
      'O usuário deve respeitar outros participantes e não publicar conteúdo ofensivo, discriminatório, fraudulento ou que viole direitos de terceiros.',
    ],
  },
  {
    title: '9. Limitação de responsabilidade',
    paragraphs: [
      'O PalpitAI busca manter o serviço disponível e funcional, mas pode haver interrupções, falhas técnicas, indisponibilidades temporárias ou alterações de funcionalidades.',
      'Na medida permitida pela legislação aplicável, o PalpitAI não se responsabiliza por perdas decorrentes do uso inadequado da plataforma, decisões tomadas com base em previsões ou eventos fora de seu controle razoável.',
    ],
  },
  {
    title: '10. Suspensão ou encerramento de conta',
    paragraphs: [
      'Contas poderão ser suspensas ou encerradas em caso de violação destes termos, uso indevido da plataforma, fraude, risco à segurança ou solicitação do próprio usuário.',
      'Quando possível, o PalpitAI poderá comunicar o usuário sobre a medida adotada.',
    ],
  },
  {
    title: '11. Alterações nos termos',
    paragraphs: [
      'Estes Termos de Uso poderão ser atualizados periodicamente para refletir mudanças no serviço, em requisitos legais ou em práticas operacionais.',
      'As alterações entram em vigor após sua publicação.',
    ],
  },
  {
    title: '12. Contato',
    paragraphs: [
      'Em caso de dúvidas sobre estes Termos de Uso, entre em contato pelo e-mail contato@palpitai.app.',
    ],
  },
];

const pageContent = {
  privacy: {
    title: 'Política de Privacidade – PalpitAI',
    updatedAt: 'Última atualização: 29 de maio de 2026',
    sections: privacySections,
  },
  terms: {
    title: 'Termos de Uso – PalpitAI',
    updatedAt: 'Última atualização: 29 de maio de 2026',
    sections: termsSections,
  },
} as const;

export default function LegalPage({ type }: { type: keyof typeof pageContent }) {
  const content = pageContent[type];

  return (
    <main className="legal-page">
      <article className="section-shell legal-article">
        <a className="legal-back-link" href={appPath()}>
          Voltar para o PalpitAI
        </a>
        <header className="legal-header">
          <p className="eyebrow">Legal</p>
          <h1>{content.title}</h1>
          <p>{content.updatedAt}</p>
        </header>

        <div className="legal-content">
          {content.sections.map((section) => (
            <section key={section.title} className="legal-section">
              <h2>{section.title}</h2>
              {section.paragraphs?.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}
              {section.bullets ? (
                <ul>
                  {section.bullets.map((bullet) => (
                    <li key={bullet}>{bullet}</li>
                  ))}
                </ul>
              ) : null}
            </section>
          ))}
        </div>
      </article>
    </main>
  );
}
