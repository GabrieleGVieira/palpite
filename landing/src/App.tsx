import FAQ from './components/FAQ';
import Features from './components/Features';
import Footer from './components/Footer';
import Header from './components/Header';
import Hero from './components/Hero';
import HowItWorks from './components/HowItWorks';
import TesterForm from './components/TesterForm';

export default function App() {
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
              <h2 id="testing-title">PalpitAI Beta para Android</h2>
              <p>
                O PalpitAI ainda está em desenvolvimento. Cadastre seu e-mail Google para receber
                acesso antecipado à versão Beta para Android.
              </p>
              <p>
                A lista será usada para organizar o Beta fechado do app de bolão online, validar o
                ranking de palpites e preparar a experiência para grupos antes do lançamento.
              </p>
            </div>
            <TesterForm />
          </div>
        </section>
        <FAQ />
      </main>
      <Footer />
    </>
  );
}
