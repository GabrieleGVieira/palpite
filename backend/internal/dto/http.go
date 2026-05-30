package dto

import "time"

type StatusResponse struct {
	App       string `json:"app"`
	Database  string `json:"database"`
	Env       string `json:"env"`
	Redis     string `json:"redis"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type SupabaseUserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	UserMetadata struct {
		FullName string `json:"full_name"`
	} `json:"user_metadata"`
}

type CreateGroupRequest struct {
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	MatchScope               string   `json:"match_scope"`
	SelectedTeams            []string `json:"selected_teams"`
	ParticipantLimit         *int     `json:"participant_limit"`
	HasUnlimitedParticipants bool     `json:"has_unlimited_participants"`
	IsPrivate                bool     `json:"is_private"`
	IsPaid                   bool     `json:"is_paid"`
	PaymentAmount            float64  `json:"payment_amount"`
	BlockPendingPredictions  bool     `json:"block_pending_predictions"`
}

type UpdateGroupRequest struct {
	Name                     string  `json:"name"`
	Description              string  `json:"description"`
	ParticipantLimit         *int    `json:"participant_limit"`
	HasUnlimitedParticipants bool    `json:"has_unlimited_participants"`
	IsPrivate                bool    `json:"is_private"`
	IsPaid                   bool    `json:"is_paid"`
	PaymentAmount            float64 `json:"payment_amount"`
	BlockPendingPredictions  bool    `json:"block_pending_predictions"`
}

type JoinGroupRequest struct {
	InviteCode string `json:"invite_code"`
}

type GroupResponse struct {
	ID                      string    `json:"id"`
	OwnerID                 string    `json:"owner_id"`
	Name                    string    `json:"name"`
	Description             string    `json:"description"`
	MatchScope              string    `json:"match_scope"`
	SelectedTeams           []string  `json:"selected_teams"`
	ParticipantLimit        *int      `json:"participant_limit"`
	IsPrivate               bool      `json:"is_private"`
	IsPaid                  bool      `json:"is_paid"`
	PaymentAmount           float64   `json:"payment_amount"`
	BlockPendingPredictions bool      `json:"block_pending_predictions"`
	InviteCode              string    `json:"invite_code"`
	CreatedAt               time.Time `json:"created_at"`
}

type GroupListItemResponse struct {
	GroupResponse
	MemberCount          int    `json:"member_count"`
	PendingRequestsCount int    `json:"pending_requests_count"`
	Role                 string `json:"role"`
	Status               string `json:"status"`
}

type JoinGroupResponse struct {
	Group            GroupListItemResponse `json:"group"`
	MembershipStatus string                `json:"membership_status"`
}

type JoinRequestResponse struct {
	RequestedAt time.Time `json:"requested_at"`
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

type GroupMemberResponse struct {
	DisplayName string    `json:"display_name"`
	JoinedAt    time.Time `json:"joined_at"`
	Role        string    `json:"role"`
	UserID      string    `json:"user_id"`
}

type ProfileResponse struct {
	AvatarURL   *string `json:"avatar_url,omitempty"`
	DisplayName string  `json:"display_name"`
}

type UpdateProfileRequest struct {
	AvatarURL   *string `json:"avatar_url"`
	DisplayName string  `json:"display_name"`
}

type GroupMemberSummaryResponse struct {
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	DisplayName string    `json:"display_name"`
	JoinedAt    time.Time `json:"joined_at"`
	Points      *int      `json:"points,omitempty"`
	Ranking     *int      `json:"ranking,omitempty"`
	Role        string    `json:"role"`
	UserID      string    `json:"user_id"`
}

type GroupMemberDetailResponse struct {
	AccuracyPercentage *float64  `json:"accuracy_percentage,omitempty"`
	AvatarURL          *string   `json:"avatar_url,omitempty"`
	CorrectPredictions *int      `json:"correct_predictions,omitempty"`
	DisplayName        string    `json:"display_name"`
	JoinedAt           time.Time `json:"joined_at"`
	Points             *int      `json:"points,omitempty"`
	PredictionsCount   *int      `json:"predictions_count,omitempty"`
	Ranking            *int      `json:"ranking,omitempty"`
	Role               string    `json:"role"`
	UserID             string    `json:"user_id"`
}

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusExempt   PaymentStatus = "exempt"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type GroupPaymentResponse struct {
	ID              string     `json:"id"`
	GroupID         string     `json:"group_id"`
	UserID          string     `json:"user_id"`
	DisplayName     string     `json:"display_name"`
	Email           *string    `json:"email"`
	AvatarURL       *string    `json:"avatar_url"`
	Status          string     `json:"status"`
	AmountExpected  float64    `json:"amount_expected"`
	AmountPaid      float64    `json:"amount_paid"`
	PaymentMethod   string     `json:"payment_method"`
	PaidAt          *time.Time `json:"paid_at"`
	MarkedByAdminID *string    `json:"marked_by_admin_id"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type UpdateGroupPaymentRequest struct {
	Status         string  `json:"status"`
	AmountExpected float64 `json:"amount_expected"`
	AmountPaid     float64 `json:"amount_paid"`
	PaymentMethod  string  `json:"payment_method"`
	Notes          string  `json:"notes"`
}

type GroupPaymentsSummaryResponse struct {
	TotalParticipants int     `json:"total_participants"`
	PaidCount         int     `json:"paid_count"`
	PendingCount      int     `json:"pending_count"`
	ExemptCount       int     `json:"exempt_count"`
	RefundedCount     int     `json:"refunded_count"`
	TotalExpected     float64 `json:"total_expected"`
	TotalPaid         float64 `json:"total_paid"`
	TotalPending      float64 `json:"total_pending"`
}

type PredictionRequest struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

type PredictionResponse struct {
	AwayScore int        `json:"away_score"`
	HomeScore int        `json:"home_score"`
	MatchID   string     `json:"match_id"`
	Points    *int       `json:"points"`
	ScoredAt  *time.Time `json:"scored_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type MatchPredictionAIResponse struct {
	Explanation   *PredictionExplanationResponse `json:"explanation,omitempty"`
	Goals         *MatchPredictionGoalsResponse  `json:"goals,omitempty"`
	MatchID       string                         `json:"match_id"`
	Probabilities MatchPredictionProbabilities   `json:"probabilities"`
	TopScores     []PredictionScoreResponse      `json:"top_scores,omitempty"`
}

type MatchPredictionProbabilities struct {
	AwayWin float64 `json:"away_win"`
	Draw    float64 `json:"draw"`
	HomeWin float64 `json:"home_win"`
}

type MatchPredictionGoalsResponse struct {
	ExpectedAwayGoals float64 `json:"expected_away_goals"`
	ExpectedHomeGoals float64 `json:"expected_home_goals"`
	MostLikelyScore   *string `json:"most_likely_score,omitempty"`
}

type PredictionScoreResponse struct {
	Probability float64 `json:"probability"`
	Score       string  `json:"score"`
}

type PredictionExplanationResponse struct {
	BetStyle    *string  `json:"bet_style,omitempty"`
	MainReasons []string `json:"main_reasons"`
	RiskAlert   *string  `json:"risk_alert"`
	Summary     string   `json:"summary"`
	UserTip     *string  `json:"user_tip"`
}

type MatchResultRequest struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

type MatchResponse struct {
	AwayTeam       string              `json:"away_team"`
	FinalAwayScore *int                `json:"final_away_score"`
	FinalHomeScore *int                `json:"final_home_score"`
	FinishedAt     *time.Time          `json:"finished_at"`
	HomeTeam       string              `json:"home_team"`
	ID             string              `json:"id"`
	KickoffAt      time.Time           `json:"kickoff_at"`
	MyPrediction   *PredictionResponse `json:"my_prediction"`
	Stage          string              `json:"stage"`
	Status         string              `json:"status"`
}

type RankingEntryResponse struct {
	Position    int    `json:"position"`
	TotalPoints int    `json:"total_points"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}
