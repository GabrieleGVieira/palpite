import FAQ from './components/FAQ';
import BetaApprovalPage from './components/BetaApprovalPage';
import Features from './components/Features';
import Footer from './components/Footer';
import Header from './components/Header';
import Hero from './components/Hero';
import HowItWorks from './components/HowItWorks';
import LegalNotice from './components/LegalNotice';
import LegalPage from './components/LegalPage';
import TesterForm from './components/TesterForm';
import { currentAppPath } from './assets';

export default function App() {
  const path = currentAppPath();

  if (path === '/privacy') {
    return (
      <>
        <Header />
        <LegalPage type="privacy" />
        <Footer />
      </>
    );
  }

  if (path === '/terms') {
    return (
      <>
        <Header />
        <LegalPage type="terms" />
        <Footer />
      </>
    );
  }

  if (path === '/account-deletion') {
    return (
      <>
        <Header />
        <LegalPage type="accountDeletion" />
        <Footer />
      </>
    );
  }

  const betaApprovalMatch = path.match(/^\/admin\/beta-test(?:er)?s\/([^/]+)\/approve\/confirm$/);
  if (betaApprovalMatch) {
    return (
      <>
        <Header />
        <BetaApprovalPage testerId={decodeURIComponent(betaApprovalMatch[1])} />
        <Footer />
      </>
    );
  }

  return (
    <>
      <Header />
      <main>
        <Hero />
        <HowItWorks />
        <Features />
        <section className="testing-section" id="teste-android" aria-labelledby="testing-title">
          <div className="section-shell testing-grid">
            <div className="testing-copy">
              <p className="eyebrow">Acesso antecipado</p>
              <h2 id="testing-title">Palpite! Beta para Android</h2>
              <p>
                O Palpite! ainda está em desenvolvimento. Cadastre seu e-mail Google para receber
                acesso antecipado à versão Beta para Android.
              </p>
              <p>
                A lista será usada para organizar o Beta fechado do app de bolão online, validar o
                ranking dos Palpiteiros e preparar a experiência para grupos antes do lançamento.
              </p>
            </div>
            <TesterForm />
            <LegalNotice className="testing-legal-notice" />
          </div>
        </section>
        <FAQ />
      </main>
      <Footer />
    </>
  );
}
