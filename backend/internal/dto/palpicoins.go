package dto

import "time"

type WalletResponse struct {
	Balance     int    `json:"balance"`
	TotalEarned int    `json:"totalEarned"`
	TotalSpent  int    `json:"totalSpent"`
	Notice      string `json:"notice"`
}

type PalpicoinTransactionResponse struct {
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"createdAt"`
	Description   string    `json:"description"`
	ID            string    `json:"id"`
	ReferenceID   *string   `json:"referenceId,omitempty"`
	ReferenceType *string   `json:"referenceType,omitempty"`
	Type          string    `json:"type"`
}

type PalpicoinTransactionPageResponse struct {
	Items  []PalpicoinTransactionResponse `json:"items"`
	Limit  int                            `json:"limit"`
	Offset int                            `json:"offset"`
	Notice string                         `json:"notice"`
}

type PalpicoinRankingEntryResponse struct {
	AvatarURL *string `json:"avatar"`
	Balance   int     `json:"saldo"`
	Name      string  `json:"nome"`
	Position  int     `json:"posicao"`
	UserID    string  `json:"userId"`
	IsCurrent bool    `json:"isCurrentUser"`
}

type PalpicoinRankingResponse struct {
	Ranking []PalpicoinRankingEntryResponse `json:"ranking"`
	Notice  string                          `json:"notice"`
}

type CreateChallengeRequest struct {
	MatchID     string `json:"matchId"`
	OpponentID  string `json:"opponentId"`
	StakeAmount int    `json:"stakeAmount"`
}

type ChallengeResponse struct {
	AcceptedAt           *time.Time `json:"acceptedAt,omitempty"`
	AwayTeam             string     `json:"awayTeam,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	CreatorPoints        *int       `json:"creatorPoints,omitempty"`
	CreatorPredictionID  *string    `json:"creatorPredictionId,omitempty"`
	CreatorUserID        string     `json:"creatorUserId"`
	FriendAvatarURL      *string    `json:"friendAvatarUrl,omitempty"`
	FriendName           string     `json:"friendName"`
	HomeTeam             string     `json:"homeTeam,omitempty"`
	ID                   string     `json:"id"`
	KickoffAt            *time.Time `json:"kickoffAt,omitempty"`
	MatchID              string     `json:"matchId"`
	OpponentPoints       *int       `json:"opponentPoints,omitempty"`
	OpponentPredictionID *string    `json:"opponentPredictionId,omitempty"`
	OpponentUserID       string     `json:"opponentUserId"`
	SettledAt            *time.Time `json:"settledAt,omitempty"`
	StakeAmount          int        `json:"stakeAmount"`
	Status               string     `json:"status"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	WinnerUserID         *string    `json:"winnerUserId,omitempty"`
}

type ChallengeListResponse struct {
	Challenges []ChallengeResponse `json:"challenges"`
	Notice     string              `json:"notice"`
}
