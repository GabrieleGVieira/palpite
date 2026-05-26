package route

import (
	"net/http"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/controller"
	predictionservice "github.com/gabrielevieira/palpitai/backend/internal/predictions/service"
	"github.com/gabrielevieira/palpitai/backend/internal/usecase"
)

type Services struct {
	Realtime controller.RealtimeService
	Redis    controller.RedisStatusService
}

func NewRouter(cfg config.Config, db usecase.Datastore, services ...Services) http.Handler {
	mux := http.NewServeMux()
	var realtimeHub controller.RealtimeService
	var redis controller.RedisStatusService
	if len(services) > 0 {
		realtimeHub = services[0].Realtime
		redis = services[0].Redis
	}
	groups := usecase.NewGroupUsecase(db)
	predictions := usecase.NewPredictionUsecase(db)
	predictionReader := predictionservice.NewPredictionReadService(db)

	mux.HandleFunc("GET /health", controller.HealthHandler(db, redis))
	mux.HandleFunc("GET /ws", controller.RealtimeHandler(cfg, db, realtimeHub))
	mux.HandleFunc("GET /api/v1/status", controller.StatusHandler(cfg, db, redis))
	mux.HandleFunc("GET /api/v1/me/score", controller.UserScoreHandler(cfg, predictions))
	mux.HandleFunc("GET /api/v1/matches/{matchID}/prediction", controller.GetMatchPredictionHandler(cfg, predictionReader))
	mux.HandleFunc("GET /api/v1/groups", controller.ListGroupsHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups", controller.CreateGroupHandler(cfg, groups))
	mux.HandleFunc("PUT /api/v1/groups/{groupID}", controller.UpdateGroupHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups/join", controller.JoinGroupHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/join-requests", controller.ListJoinRequestsHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups/{groupID}/join-requests/{userID}/approve", controller.ApproveJoinRequestHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/matches", controller.ListGroupMatchesHandler(cfg, predictions))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/ranking", controller.GroupRankingHandler(cfg, predictions))
	mux.HandleFunc("PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction", controller.SavePredictionHandler(cfg, predictions))
	mux.HandleFunc("PUT /api/v1/matches/{matchID}/result", controller.SaveMatchResultHandler(cfg, predictions, realtimeHub))

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
