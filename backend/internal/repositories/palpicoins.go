package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
)

const PalpicoinNotice = "Palpicoins são moedas virtuais sem valor monetário."

func EnsureWallet(ctx context.Context, db Querier, userID string) error {
	_, err := db.Exec(ctx, `select ensure_user_wallet($1)`, userID)
	return err
}

func WalletByUserID(ctx context.Context, db Querier, userID string) (dto.WalletResponse, error) {
	if err := EnsureWallet(ctx, db, userID); err != nil {
		return dto.WalletResponse{}, err
	}

	var wallet dto.WalletResponse
	err := db.QueryRow(ctx, `
		select balance, total_earned, total_spent
		from user_wallets
		where user_id = $1
	`, userID).Scan(&wallet.Balance, &wallet.TotalEarned, &wallet.TotalSpent)
	wallet.Notice = PalpicoinNotice
	return wallet, err
}

func ListPalpicoinTransactions(ctx context.Context, db Querier, userID string, limit int, offset int) ([]dto.PalpicoinTransactionResponse, error) {
	if err := EnsureWallet(ctx, db, userID); err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, `
		select id::text, amount, type, description, reference_type, reference_id::text, created_at
		from palpicoin_transactions
		where user_id = $1
		order by created_at desc
		limit $2 offset $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []dto.PalpicoinTransactionResponse{}
	for rows.Next() {
		var item dto.PalpicoinTransactionResponse
		if err := rows.Scan(&item.ID, &item.Amount, &item.Type, &item.Description, &item.ReferenceType, &item.ReferenceID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func CreditWallet(ctx context.Context, db Querier, userID string, amount int, transactionType domain.PalpicoinTransactionType, description string, referenceType *string, referenceID *string) (bool, error) {
	if amount <= 0 {
		return false, errors.New("credit amount must be positive")
	}
	if err := EnsureWallet(ctx, db, userID); err != nil {
		return false, err
	}
	return moveWalletBalance(ctx, db, userID, amount, transactionType, description, referenceType, referenceID)
}

func DebitWallet(ctx context.Context, db Querier, userID string, amount int, transactionType domain.PalpicoinTransactionType, description string, referenceType *string, referenceID *string) (bool, error) {
	if amount <= 0 {
		return false, errors.New("debit amount must be positive")
	}
	if err := EnsureWallet(ctx, db, userID); err != nil {
		return false, err
	}
	return moveWalletBalance(ctx, db, userID, -amount, transactionType, description, referenceType, referenceID)
}

func moveWalletBalance(ctx context.Context, db Querier, userID string, signedAmount int, transactionType domain.PalpicoinTransactionType, description string, referenceType *string, referenceID *string) (bool, error) {
	var transactionID string
	err := db.QueryRow(ctx, `
		with inserted as (
			insert into palpicoin_transactions (
				user_id,
				amount,
				type,
				description,
				reference_type,
				reference_id
			)
			select $1, $2, $3, $4, $5, $6::uuid
			where $2 > 0 or exists (
				select 1
				from user_wallets
				where user_id = $1
					and balance >= abs($2)
			)
			on conflict (user_id, type, reference_type, reference_id)
				where reference_type is not null and reference_id is not null
			do nothing
			returning id
		),
		updated_wallet as (
			update user_wallets
			set
				balance = balance + $2,
				total_earned = total_earned + case when $2 > 0 then $2 else 0 end,
				total_spent = total_spent + case when $2 < 0 then abs($2) else 0 end,
				updated_at = now()
			where user_id = $1
				and exists (select 1 from inserted)
				and balance + $2 >= 0
			returning user_id
		)
		select coalesce((select id::text from inserted), '')
	`, userID, signedAmount, transactionType, description, referenceType, referenceID).Scan(&transactionID)
	if err != nil {
		return false, err
	}
	return transactionID != "", nil
}

func PalpicoinRanking(ctx context.Context, db Querier, currentUserID string) ([]dto.PalpicoinRankingEntryResponse, error) {
	rows, err := db.Query(ctx, `
		with profiles as (
			select distinct on (gm.user_id)
				gm.user_id,
				gm.display_name,
				gm.avatar_url
			from group_members gm
			where gm.status = 'active'
			order by gm.user_id, gm.joined_at desc
		),
		ranked as (
			select
				uw.user_id,
				uw.balance,
				coalesce(p.display_name, '') as display_name,
				p.avatar_url,
				rank() over (order by uw.balance desc, coalesce(p.display_name, '') asc)::int as position
			from user_wallets uw
			left join profiles p on p.user_id = uw.user_id
			left join user_social_settings uss on uss.user_id = uw.user_id
			where uw.user_id = $1::uuid
				or coalesce(uss.is_public_profile, true) = true
				or exists (
					select 1
					from friendships f
					where f.status = 'ACCEPTED'
						and least(f.requester_user_id, f.addressee_user_id) = least($1::uuid, uw.user_id)
						and greatest(f.requester_user_id, f.addressee_user_id) = greatest($1::uuid, uw.user_id)
				)
		)
		select position, user_id::text, coalesce(nullif(display_name, ''), 'Palpiteiro'), avatar_url, balance, user_id = $1::uuid
		from ranked
		order by position asc, display_name asc
		limit 100
	`, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ranking := []dto.PalpicoinRankingEntryResponse{}
	for rows.Next() {
		var entry dto.PalpicoinRankingEntryResponse
		if err := rows.Scan(&entry.Position, &entry.UserID, &entry.Name, &entry.AvatarURL, &entry.Balance, &entry.IsCurrent); err != nil {
			return nil, err
		}
		if strings.TrimSpace(entry.Name) == "" {
			entry.Name = "Palpiteiro"
		}
		ranking = append(ranking, entry)
	}
	return ranking, rows.Err()
}

func AcceptedFriendshipExists(ctx context.Context, db Querier, userID string, otherUserID string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from friendships
			where status = 'ACCEPTED'
				and least(requester_user_id, addressee_user_id) = least($1::uuid, $2::uuid)
				and greatest(requester_user_id, addressee_user_id) = greatest($1::uuid, $2::uuid)
		)
	`, userID, otherUserID).Scan(&exists)
	return exists, err
}

func ScheduledMatchExists(ctx context.Context, db Querier, matchID string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from world_cup_matches
			where id = $1
				and status = 'scheduled'
				and kickoff_at > now()
		)
	`, matchID).Scan(&exists)
	return exists, err
}
