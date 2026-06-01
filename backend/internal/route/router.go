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
	accounts := usecase.NewAccountUsecase(db)
	groups := usecase.NewGroupUsecase(db)
	friends := usecase.NewFriendsUsecase(db)
	feed := usecase.NewFeedUsecase(db)
	predictions := usecase.NewPredictionUsecase(db)
	predictionReader := predictionservice.NewPredictionReadService(db)
	wallet := usecase.NewWalletUsecase(db)
	challenges := usecase.NewChallengeUsecase(db)

	mux.HandleFunc("GET /health", controller.HealthHandler(db, redis))
	mux.HandleFunc("GET /ws", controller.RealtimeHandler(cfg, db, realtimeHub))
	mux.HandleFunc("GET /api/v1/status", controller.StatusHandler(cfg, db, redis))
	mux.HandleFunc("DELETE /api/v1/me", controller.DeleteAccountHandler(cfg, accounts))
	mux.HandleFunc("GET /api/v1/me/profile", controller.GetProfileHandler(cfg, accounts))
	mux.HandleFunc("PATCH /api/v1/me/profile", controller.UpdateProfileHandler(cfg, accounts))
	mux.HandleFunc("GET /api/v1/me/score", controller.UserScoreHandler(cfg, predictions))
	mux.HandleFunc("GET /api/v1/me/wallet", controller.GetWalletHandler(cfg, wallet))
	mux.HandleFunc("GET /api/v1/me/wallet/transactions", controller.ListWalletTransactionsHandler(cfg, wallet))
	mux.HandleFunc("GET /api/v1/rankings/palpicoins", controller.PalpicoinRankingHandler(cfg, wallet))
	mux.HandleFunc("POST /api/v1/friends/request", controller.CreateFriendRequestHandler(cfg, friends))
	mux.HandleFunc("POST /api/v1/friends/{id}/accept", controller.AcceptFriendRequestHandler(cfg, friends))
	mux.HandleFunc("POST /api/v1/friends/{id}/decline", controller.DeclineFriendRequestHandler(cfg, friends))
	mux.HandleFunc("DELETE /api/v1/friends/{id}", controller.DeleteFriendshipHandler(cfg, friends))
	mux.HandleFunc("GET /api/v1/friends", controller.ListFriendsHandler(cfg, friends))
	mux.HandleFunc("GET /api/v1/friends/requests", controller.ListFriendRequestsHandler(cfg, friends))
	mux.HandleFunc("GET /api/v1/users/search", controller.SearchUsersHandler(cfg, friends))
	mux.HandleFunc("GET /api/v1/users/{id}/profile", controller.PublicProfileHandler(cfg, friends))
	mux.HandleFunc("POST /api/v1/challenges", controller.CreateChallengeHandler(cfg, challenges))
	mux.HandleFunc("POST /api/v1/challenges/{id}/accept", controller.AcceptChallengeHandler(cfg, challenges))
	mux.HandleFunc("POST /api/v1/challenges/{id}/decline", controller.DeclineChallengeHandler(cfg, challenges))
	mux.HandleFunc("POST /api/v1/challenges/{id}/cancel", controller.CancelChallengeHandler(cfg, challenges))
	mux.HandleFunc("GET /api/v1/challenges", controller.ListChallengesHandler(cfg, challenges))
	mux.HandleFunc("GET /api/v1/challenges/{id}", controller.GetChallengeHandler(cfg, challenges))
	mux.HandleFunc("GET /api/v1/matches/{matchID}/prediction", controller.GetMatchPredictionHandler(cfg, predictionReader))
	mux.HandleFunc("GET /api/v1/groups", controller.ListGroupsHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups", controller.CreateGroupHandler(cfg, groups))
	mux.HandleFunc("PUT /api/v1/groups/{groupID}", controller.UpdateGroupHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups/join", controller.JoinGroupHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/join-requests", controller.ListJoinRequestsHandler(cfg, groups))
	mux.HandleFunc("POST /api/v1/groups/{groupID}/join-requests/{userID}/approve", controller.ApproveJoinRequestHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/members", controller.ListGroupMembersHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/members/{userID}", controller.GroupMemberDetailHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/feed", controller.GroupFeedHandler(cfg, feed))
	mux.HandleFunc("POST /api/v1/groups/{groupID}/feed/{eventID}/reaction", controller.ReactToFeedEventHandler(cfg, feed))
	mux.HandleFunc("DELETE /api/v1/groups/{groupID}/feed/{eventID}/reaction", controller.DeleteFeedReactionHandler(cfg, feed))
	mux.HandleFunc("POST /api/v1/groups/{groupID}/members/{userID}/transfer-ownership", controller.TransferGroupOwnershipHandler(cfg, groups))
	mux.HandleFunc("DELETE /api/v1/groups/{groupID}/members/{userID}", controller.RemoveGroupMemberHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/payments", controller.ListGroupPaymentsHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/payments/summary", controller.GroupPaymentsSummaryHandler(cfg, groups))
	mux.HandleFunc("PATCH /api/v1/groups/{groupID}/payments/{userID}", controller.UpdateGroupPaymentHandler(cfg, groups))
	mux.HandleFunc("DELETE /api/v1/groups/{groupID}/membership", controller.LeaveGroupHandler(cfg, groups))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/matches", controller.ListGroupMatchesHandler(cfg, predictions))
	mux.HandleFunc("GET /api/v1/groups/{groupID}/ranking", controller.GroupRankingHandler(cfg, predictions))
	mux.HandleFunc("PUT /api/v1/groups/{groupID}/matches/{matchID}/prediction", controller.SavePredictionHandler(cfg, predictions))
	mux.HandleFunc("PUT /api/v1/matches/{matchID}/result", controller.SaveMatchResultHandler(cfg, predictions, realtimeHub))

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
