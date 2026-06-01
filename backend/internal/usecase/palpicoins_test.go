package usecase

import (
	"testing"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
)

func TestPalpicoinRewardsForWinnerHit(t *testing.T) {
	rewards := PalpicoinRewardsForPrediction(2, 1, 3, 0)

	if len(rewards) != 1 {
		t.Fatalf("expected one reward, got %#v", rewards)
	}
	if rewards[0].Amount != 10 || rewards[0].Type != domain.PalpicoinTransactionMatchWinnerHit {
		t.Fatalf("unexpected reward: %#v", rewards[0])
	}
}

func TestPalpicoinRewardsForDrawHit(t *testing.T) {
	rewards := PalpicoinRewardsForPrediction(1, 1, 2, 2)

	if len(rewards) != 1 {
		t.Fatalf("expected one reward, got %#v", rewards)
	}
	if rewards[0].Amount != 15 || rewards[0].Description != "Acerto de empate" {
		t.Fatalf("unexpected draw reward: %#v", rewards[0])
	}
}

func TestPalpicoinRewardsForExactScoreHit(t *testing.T) {
	rewards := PalpicoinRewardsForPrediction(2, 1, 2, 1)

	if len(rewards) != 2 {
		t.Fatalf("expected exact score and winner rewards, got %#v", rewards)
	}
	if rewards[0].Amount != 50 || rewards[0].Type != domain.PalpicoinTransactionExactScoreHit {
		t.Fatalf("unexpected exact score reward: %#v", rewards[0])
	}
	if rewards[1].Amount != 10 || rewards[1].Type != domain.PalpicoinTransactionMatchWinnerHit {
		t.Fatalf("unexpected winner reward: %#v", rewards[1])
	}
}

func TestChallengePayoutsForWinner(t *testing.T) {
	winnerID := "user-1"
	payouts := ChallengePayouts(domain.PalpicoinChallenge{
		CreatorUserID:  "user-1",
		OpponentUserID: "user-2",
		StakeAmount:    200,
		WinnerUserID:   &winnerID,
	})

	if len(payouts) != 1 {
		t.Fatalf("expected one payout, got %#v", payouts)
	}
	if payouts[0].UserID != winnerID || payouts[0].Amount != 400 || payouts[0].Type != domain.PalpicoinTransactionChallengeWin {
		t.Fatalf("unexpected winner payout: %#v", payouts[0])
	}
}

func TestChallengePayoutsForTie(t *testing.T) {
	payouts := ChallengePayouts(domain.PalpicoinChallenge{
		CreatorUserID:  "user-1",
		OpponentUserID: "user-2",
		StakeAmount:    150,
	})

	if len(payouts) != 2 {
		t.Fatalf("expected two refunds, got %#v", payouts)
	}
	for _, payout := range payouts {
		if payout.Amount != 150 || payout.Type != domain.PalpicoinTransactionChallengeRefund {
			t.Fatalf("unexpected tie payout: %#v", payout)
		}
	}
}

func TestCanDebitWalletBalance(t *testing.T) {
	tests := []struct {
		name    string
		balance int
		amount  int
		want    bool
	}{
		{name: "exact balance", balance: 100, amount: 100, want: true},
		{name: "available balance", balance: 150, amount: 100, want: true},
		{name: "negative balance blocked", balance: 50, amount: 100, want: false},
		{name: "invalid amount blocked", balance: 100, amount: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanDebitWalletBalance(tt.balance, tt.amount); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
