package domain

import "time"

type PalpicoinTransactionType string

const (
	PalpicoinTransactionSignupBonus     PalpicoinTransactionType = "SIGNUP_BONUS"
	PalpicoinTransactionMatchWinnerHit  PalpicoinTransactionType = "MATCH_WINNER_HIT"
	PalpicoinTransactionExactScoreHit   PalpicoinTransactionType = "EXACT_SCORE_HIT"
	PalpicoinTransactionRoundTop3       PalpicoinTransactionType = "ROUND_TOP_3"
	PalpicoinTransactionChallengeStake  PalpicoinTransactionType = "CHALLENGE_STAKE"
	PalpicoinTransactionChallengeWin    PalpicoinTransactionType = "CHALLENGE_WIN"
	PalpicoinTransactionChallengeRefund PalpicoinTransactionType = "CHALLENGE_REFUND"
)

type Wallet struct {
	ID          string
	UserID      string
	Balance     int
	TotalEarned int
	TotalSpent  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PalpicoinTransaction struct {
	ID            string
	UserID        string
	Amount        int
	Type          PalpicoinTransactionType
	Description   string
	ReferenceType *string
	ReferenceID   *string
	CreatedAt     time.Time
}

type ChallengeStatus string

const (
	ChallengeStatusPending   ChallengeStatus = "PENDING"
	ChallengeStatusAccepted  ChallengeStatus = "ACCEPTED"
	ChallengeStatusDeclined  ChallengeStatus = "DECLINED"
	ChallengeStatusCancelled ChallengeStatus = "CANCELLED"
	ChallengeStatusSettled   ChallengeStatus = "SETTLED"
)

type PalpicoinChallenge struct {
	ID                   string
	CreatorUserID        string
	OpponentUserID       string
	MatchID              string
	StakeAmount          int
	CreatorPredictionID  *string
	OpponentPredictionID *string
	CreatorPoints        *int
	OpponentPoints       *int
	WinnerUserID         *string
	Status               ChallengeStatus
	CreatedAt            time.Time
	AcceptedAt           *time.Time
	SettledAt            *time.Time
	UpdatedAt            time.Time
}

const (
	SocialEventChallengeCreated  SocialEventType = "CHALLENGE_CREATED"
	SocialEventChallengeAccepted SocialEventType = "CHALLENGE_ACCEPTED"
	SocialEventChallengeWon      SocialEventType = "CHALLENGE_WON"
)
