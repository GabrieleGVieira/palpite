package matchsync

import (
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

const (
	// defaultRequestTimeout limita chamadas externas para o worker nao travar indefinidamente.
	defaultRequestTimeout = 10 * time.Second
	// livePollInterval define a frequencia de atualizacao para partidas ao vivo.
	livePollInterval = 30 * time.Second
	// liveRecentWindow define uma margem para capturar jogos que sairam de live entre polls.
	liveRecentWindow = 5 * time.Hour
	// rateLimitGap mantem um espacamento minimo entre chamadas ao football-data.
	rateLimitGap = 6 * time.Second
	// todayPollInterval define a frequencia para revisar jogos do dia.
	todayPollInterval = 5 * time.Minute
	// upcomingPollInterval define a frequencia para buscar jogos futuros.
	upcomingPollInterval = time.Hour
	// upcomingWindow define quantos dias futuros entram na sincronizacao de proximas partidas.
	upcomingWindow = 30 * 24 * time.Hour
)

// datastore reutiliza a interface de banco dos repositorios para facilitar testes do Syncer.
type datastore = repositories.Datastore

// syncKind identifica qual recorte de partidas sera consultado no provedor.
type syncKind string

const (
	// syncLive representa partidas em andamento ou recem-finalizadas.
	syncLive syncKind = "live"
	// syncToday representa partidas marcadas para o dia atual.
	syncToday syncKind = "today"
	// syncUpcoming representa partidas futuras dentro da janela configurada.
	syncUpcoming syncKind = "upcoming"
)
