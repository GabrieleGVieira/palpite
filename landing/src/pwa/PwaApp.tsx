import { useEffect, useMemo, useState, type FormEvent } from 'react';
import {
  getMatchPrediction,
  isSupabaseConfigured,
  listGroupMatches,
  listGroupRanking,
  listGroups,
  savePrediction,
  supabase,
} from './api';
import type { AuthMode, Group, GroupMatch, MatchPrediction, RankingEntry } from './types';

type View = 'groups' | 'detail' | 'matches' | 'prediction' | 'ranking' | 'ai';

const demoGroups: Group[] = [
  {
    description: 'Bolão da rodada para manter a resenha em dia.',
    id: 'demo',
    invite_code: 'PALPITE',
    member_count: 12,
    name: 'Copa dos Amigos',
    pending_requests_count: 0,
    role: 'member',
    status: 'active',
  },
];

const demoMatches: GroupMatch[] = [
  {
    away_team: 'Argentina',
    final_away_score: null,
    final_home_score: null,
    finished_at: null,
    home_team: 'Brasil',
    id: 'demo-match-1',
    kickoff_at: new Date(Date.now() + 36 * 60 * 60 * 1000).toISOString(),
    my_prediction: null,
    stage: 'Rodada 1',
    status: 'scheduled',
  },
  {
    away_team: 'Espanha',
    final_away_score: null,
    final_home_score: null,
    finished_at: null,
    home_team: 'França',
    id: 'demo-match-2',
    kickoff_at: new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString(),
    my_prediction: { away_score: 1, home_score: 2, match_id: 'demo-match-2', points: null, scored_at: null, updated_at: new Date().toISOString() },
    stage: 'Rodada 1',
    status: 'scheduled',
  },
];

const demoRanking: RankingEntry[] = [
  { display_name: 'Você', position: 1, total_points: 18, user_id: 'demo-user' },
  { display_name: 'Camisa 10', position: 2, total_points: 15, user_id: 'demo-2' },
  { display_name: 'Zagueiro raiz', position: 3, total_points: 12, user_id: 'demo-3' },
];

export default function PwaApp() {
  const [authMode, setAuthMode] = useState<AuthMode>('login');
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [isSignedIn, setIsSignedIn] = useState(false);
  const [view, setView] = useState<View>('groups');
  const [groups, setGroups] = useState<Group[]>(demoGroups);
  const [matches, setMatches] = useState<GroupMatch[]>(demoMatches);
  const [ranking, setRanking] = useState<RankingEntry[]>(demoRanking);
  const [selectedGroupID, setSelectedGroupID] = useState(demoGroups[0].id);
  const [selectedMatchID, setSelectedMatchID] = useState(demoMatches[0].id);
  const [homeScore, setHomeScore] = useState('1');
  const [awayScore, setAwayScore] = useState('0');
  const [aiPrediction, setAiPrediction] = useState<MatchPrediction | null>(null);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  const selectedGroup = useMemo(
    () => groups.find((group) => group.id === selectedGroupID) ?? groups[0],
    [groups, selectedGroupID],
  );
  const selectedMatch = useMemo(
    () => matches.find((match) => match.id === selectedMatchID) ?? matches[0],
    [matches, selectedMatchID],
  );

  useEffect(() => {
    document.title = 'Palpite! App';
    const robots = document.querySelector<HTMLMetaElement>('meta[name="robots"]');
    robots?.setAttribute('content', 'noindex, nofollow');

    if (!isSupabaseConfigured) {
      setMessage('Configure VITE_SUPABASE_URL e VITE_SUPABASE_ANON_KEY para usar login real.');
      return;
    }

    supabase.auth.getSession().then(({ data }) => {
      const signed = Boolean(data.session);
      setIsSignedIn(signed);

      if (signed) {
        void loadGroups();
      }
    });

    const { data } = supabase.auth.onAuthStateChange((_event, session) => {
      setIsSignedIn(Boolean(session));
    });

    return () => data.subscription.unsubscribe();
  }, []);

  useEffect(() => {
    const match = matches.find((item) => item.id === selectedMatchID);

    if (match?.my_prediction) {
      setHomeScore(String(match.my_prediction.home_score));
      setAwayScore(String(match.my_prediction.away_score));
    }
  }, [matches, selectedMatchID]);

  async function handleAuth(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!isSupabaseConfigured) {
      setMessage('Modo demonstração ativo. Publique as variáveis do Supabase para login real.');
      setIsSignedIn(true);
      return;
    }

    setLoading(true);
    setMessage('');

    try {
      if (authMode === 'login') {
        const { error } = await supabase.auth.signInWithPassword({ email: email.trim(), password });
        if (error) throw error;
      } else {
        const { error } = await supabase.auth.signUp({
          email: email.trim(),
          options: { data: { full_name: name.trim() } },
          password,
        });
        if (error) throw error;
        setMessage('Cadastro enviado. Se precisar confirmar e-mail, confira sua caixa de entrada.');
      }

      setIsSignedIn(true);
      await loadGroups();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível entrar agora.');
    } finally {
      setLoading(false);
    }
  }

  async function loadGroups() {
    setLoading(true);

    try {
      const nextGroups = await listGroups();
      setGroups(nextGroups.length ? nextGroups : demoGroups);
      setSelectedGroupID(nextGroups[0]?.id ?? demoGroups[0].id);

      if (nextGroups[0]) {
        await loadGroupData(nextGroups[0].id);
      }
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível carregar os bolões.');
    } finally {
      setLoading(false);
    }
  }

  async function loadGroupData(groupID: string) {
    setLoading(true);
    setSelectedGroupID(groupID);

    try {
      const [nextMatches, nextRanking] = await Promise.all([
        listGroupMatches(groupID),
        listGroupRanking(groupID),
      ]);
      setMatches(nextMatches.length ? nextMatches : demoMatches);
      setRanking(nextRanking.length ? nextRanking : demoRanking);
      setSelectedMatchID(nextMatches[0]?.id ?? demoMatches[0].id);
      setMessage('');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível carregar este bolão.');
    } finally {
      setLoading(false);
    }
  }

  async function handleSavePrediction(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!selectedGroup || !selectedMatch) return;

    setLoading(true);

    try {
      const payload = { away_score: Number(awayScore), home_score: Number(homeScore) };

      if (isSignedIn && selectedGroup.id !== 'demo') {
        await savePrediction(selectedGroup.id, selectedMatch.id, payload);
      }

      setMatches((current) =>
        current.map((match) =>
          match.id === selectedMatch.id
            ? {
                ...match,
                my_prediction: {
                  ...payload,
                  match_id: match.id,
                  points: null,
                  scored_at: null,
                  updated_at: new Date().toISOString(),
                },
              }
            : match,
        ),
      );
      setMessage('Palpite salvo. Agora é secar com categoria.');
      setView('matches');
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível salvar o palpite.');
    } finally {
      setLoading(false);
    }
  }

  async function openAi(match: GroupMatch) {
    setSelectedMatchID(match.id);
    setAiPrediction(null);
    setView('ai');

    if (!isSignedIn || match.id.startsWith('demo')) {
      setAiPrediction({
        confidence: 0.62,
        explanation: 'A PalpitAI tende ao mandante por fase recente e mando, mas vê jogo parelho.',
        match_id: match.id,
        predicted_away_score: 1,
        predicted_home_score: 2,
        probabilities: { away_win: 0.23, draw: 0.29, home_win: 0.48 },
        recommended_prediction: `${match.home_team} 2 x 1 ${match.away_team}`,
        summary: 'Sugestão moderada, boa para quem quer fugir do empate.',
      });
      return;
    }

    try {
      setLoading(true);
      setAiPrediction(await getMatchPrediction(match.id));
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Não foi possível abrir a PalpitAI.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="pwa-shell">
      <section className="pwa-phone" aria-label="Palpite PWA">
        <header className="pwa-topbar">
          <div>
            <span className="pwa-kicker">Palpite!</span>
            <h1>Bolão na mão</h1>
          </div>
          {isSignedIn ? (
            <button className="pwa-icon-button" type="button" onClick={() => supabase.auth.signOut()}>
              Sair
            </button>
          ) : null}
        </header>

        {!isSignedIn ? (
          <AuthScreen
            authMode={authMode}
            email={email}
            isConfigured={isSupabaseConfigured}
            loading={loading}
            message={message}
            name={name}
            password={password}
            setAuthMode={setAuthMode}
            setEmail={setEmail}
            setName={setName}
            setPassword={setPassword}
            onSubmit={handleAuth}
          />
        ) : (
          <>
            {message ? <p className="pwa-message">{message}</p> : null}
            {view === 'groups' ? (
              <GroupsScreen groups={groups} loading={loading} onOpen={(groupID) => { void loadGroupData(groupID); setView('detail'); }} />
            ) : null}
            {view === 'detail' && selectedGroup ? (
              <GroupDetailScreen group={selectedGroup} onRanking={() => setView('ranking')} onMatches={() => setView('matches')} />
            ) : null}
            {view === 'matches' ? (
              <MatchesScreen matches={matches} onAi={openAi} onPredict={(match) => { setSelectedMatchID(match.id); setView('prediction'); }} />
            ) : null}
            {view === 'prediction' && selectedMatch ? (
              <PredictionScreen
                awayScore={awayScore}
                homeScore={homeScore}
                loading={loading}
                match={selectedMatch}
                setAwayScore={setAwayScore}
                setHomeScore={setHomeScore}
                onSubmit={handleSavePrediction}
              />
            ) : null}
            {view === 'ranking' ? <RankingScreen ranking={ranking} /> : null}
            {view === 'ai' && selectedMatch ? (
              <AiScreen loading={loading} match={selectedMatch} prediction={aiPrediction} />
            ) : null}
            <nav className="pwa-tabbar" aria-label="Navegação do app">
              <button type="button" aria-current={view === 'groups'} onClick={() => setView('groups')}>Bolões</button>
              <button type="button" aria-current={view === 'matches'} onClick={() => setView('matches')}>Jogos</button>
              <button type="button" aria-current={view === 'ranking'} onClick={() => setView('ranking')}>Ranking</button>
              <button type="button" aria-current={view === 'ai'} onClick={() => selectedMatch && void openAi(selectedMatch)}>PalpitAI</button>
            </nav>
          </>
        )}
      </section>
    </main>
  );
}

function AuthScreen(props: {
  authMode: AuthMode;
  email: string;
  isConfigured: boolean;
  loading: boolean;
  message: string;
  name: string;
  password: string;
  setAuthMode: (mode: AuthMode) => void;
  setEmail: (value: string) => void;
  setName: (value: string) => void;
  setPassword: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <section className="pwa-panel">
      <div className="pwa-segmented">
        <button type="button" aria-pressed={props.authMode === 'login'} onClick={() => props.setAuthMode('login')}>Entrar</button>
        <button type="button" aria-pressed={props.authMode === 'signup'} onClick={() => props.setAuthMode('signup')}>Cadastrar</button>
      </div>
      <form className="pwa-form" onSubmit={props.onSubmit}>
        {props.authMode === 'signup' ? (
          <label>
            Nome
            <input autoComplete="name" value={props.name} onChange={(event) => props.setName(event.target.value)} />
          </label>
        ) : null}
        <label>
          E-mail
          <input autoComplete="email" inputMode="email" type="email" value={props.email} onChange={(event) => props.setEmail(event.target.value)} />
        </label>
        <label>
          Senha
          <input autoComplete={props.authMode === 'login' ? 'current-password' : 'new-password'} type="password" value={props.password} onChange={(event) => props.setPassword(event.target.value)} />
        </label>
        <button className="pwa-primary" disabled={props.loading} type="submit">
          {props.authMode === 'login' ? 'Entrar no bolão' : 'Criar conta'}
        </button>
      </form>
      {props.message ? <p className="pwa-message-inline">{props.message}</p> : null}
      {!props.isConfigured ? (
        <p className="pwa-muted">Modo demonstração ativo até configurar Supabase no deploy. Toque em Entrar para navegar.</p>
      ) : null}
    </section>
  );
}

function GroupsScreen({ groups, loading, onOpen }: { groups: Group[]; loading: boolean; onOpen: (groupID: string) => void }) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="Meus bolões" title="Escolha a arena" />
      {loading ? <p className="pwa-muted">Carregando...</p> : null}
      {groups.map((group) => (
        <button className="pwa-list-card" key={group.id} type="button" onClick={() => onOpen(group.id)}>
          <strong>{group.name}</strong>
          <span>{group.description || `${group.member_count} Palpiteiros na disputa`}</span>
          <small>{group.member_count} membros · convite {group.invite_code}</small>
        </button>
      ))}
    </section>
  );
}

function GroupDetailScreen({ group, onMatches, onRanking }: { group: Group; onMatches: () => void; onRanking: () => void }) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="Detalhe do bolão" title={group.name} />
      <div className="pwa-score-card">
        <span>{group.member_count}</span>
        <p>Palpiteiros disputando ponto a ponto.</p>
      </div>
      <div className="pwa-actions-grid">
        <button type="button" onClick={onMatches}>Ver jogos</button>
        <button type="button" onClick={onRanking}>Abrir ranking</button>
      </div>
    </section>
  );
}

function MatchesScreen({ matches, onAi, onPredict }: { matches: GroupMatch[]; onAi: (match: GroupMatch) => void; onPredict: (match: GroupMatch) => void }) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="Lista de jogos" title="Rodada quente" />
      {matches.map((match) => (
        <article className="pwa-match" key={match.id}>
          <small>{formatDate(match.kickoff_at)} · {match.stage}</small>
          <div className="pwa-versus">
            <strong>{match.home_team}</strong>
            <span>x</span>
            <strong>{match.away_team}</strong>
          </div>
          <p>{match.my_prediction ? `Seu palpite: ${match.my_prediction.home_score} x ${match.my_prediction.away_score}` : 'Palpite aberto'}</p>
          <div className="pwa-match-actions">
            <button type="button" onClick={() => onPredict(match)}>Palpitar</button>
            <button type="button" onClick={() => onAi(match)}>PalpitAI</button>
          </div>
        </article>
      ))}
    </section>
  );
}

function PredictionScreen(props: {
  awayScore: string;
  homeScore: string;
  loading: boolean;
  match: GroupMatch;
  setAwayScore: (value: string) => void;
  setHomeScore: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="Criar/editar palpite" title={`${props.match.home_team} x ${props.match.away_team}`} />
      <form className="pwa-prediction-form" onSubmit={props.onSubmit}>
        <label>
          {props.match.home_team}
          <input inputMode="numeric" min="0" type="number" value={props.homeScore} onChange={(event) => props.setHomeScore(event.target.value)} />
        </label>
        <label>
          {props.match.away_team}
          <input inputMode="numeric" min="0" type="number" value={props.awayScore} onChange={(event) => props.setAwayScore(event.target.value)} />
        </label>
        <button className="pwa-primary" disabled={props.loading} type="submit">Salvar palpite</button>
      </form>
    </section>
  );
}

function RankingScreen({ ranking }: { ranking: RankingEntry[] }) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="Ranking" title="Mesa dos líderes" />
      {ranking.map((entry) => (
        <div className="pwa-ranking-row" key={entry.user_id}>
          <span>{entry.position}</span>
          <strong>{entry.display_name}</strong>
          <small>{entry.total_points} pts</small>
        </div>
      ))}
    </section>
  );
}

function AiScreen({ loading, match, prediction }: { loading: boolean; match: GroupMatch; prediction: MatchPrediction | null }) {
  return (
    <section className="pwa-stack">
      <ScreenTitle eyebrow="PalpitAI" title={`${match.home_team} x ${match.away_team}`} />
      <div className="pwa-ai-card">
        {loading ? <p>Carregando análise...</p> : null}
        <strong>{prediction?.recommended_prediction ?? 'Análise ainda indisponível'}</strong>
        <p>{prediction?.summary ?? prediction?.explanation ?? 'Quando a API retornar uma análise para este jogo, ela aparece aqui.'}</p>
        {prediction?.probabilities ? (
          <div className="pwa-probabilities">
            <span>Mandante {formatPercent(prediction.probabilities.home_win)}</span>
            <span>Empate {formatPercent(prediction.probabilities.draw)}</span>
            <span>Visitante {formatPercent(prediction.probabilities.away_win)}</span>
          </div>
        ) : null}
      </div>
    </section>
  );
}

function ScreenTitle({ eyebrow, title }: { eyebrow: string; title: string }) {
  return (
    <div className="pwa-screen-title">
      <p>{eyebrow}</p>
      <h2>{title}</h2>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('pt-BR', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: 'short',
  }).format(new Date(value));
}

function formatPercent(value?: number | null) {
  if (value == null) return '-';
  return `${Math.round(value * 100)}%`;
}
