package repositories

import (
	"context"
	"errors"

	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

func EnsureGroupOwner(ctx context.Context, db Querier, ownerID string, groupID string) error {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from groups
			where id = $1 and owner_id = $2
		)
	`, groupID, ownerID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	return nil
}

func EnsureActiveGroupMember(ctx context.Context, db Querier, groupID string, userID string) error {
	var exists bool
	err := db.QueryRow(ctx, `
		select exists (
			select 1
			from group_members
			where group_id = $1
				and user_id = $2
				and status = 'active'
		)
	`, groupID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	return nil
}

func InsertMissingPaymentsForPaidGroup(ctx context.Context, db Querier, groupID string) error {
	_, err := db.Exec(ctx, `
		insert into group_payments (
			group_id,
			user_id,
			status,
			amount_expected
		)
		select
			gm.group_id,
			gm.user_id,
			case when gm.role = 'owner' then 'exempt' else 'pending' end,
			case when gm.role = 'owner' then 0 else g.payment_amount end
		from group_members gm
		join groups g on g.id = gm.group_id
		where gm.group_id = $1
			and gm.status = 'active'
			and g.is_paid = true
		on conflict (group_id, user_id) do update
		set
			amount_expected = case
				when group_payments.status = 'pending' then excluded.amount_expected
				else group_payments.amount_expected
			end,
			updated_at = case
				when group_payments.status = 'pending' then now()
				else group_payments.updated_at
			end
	`, groupID)

	return err
}

func InsertPaymentForMemberIfPaidGroup(ctx context.Context, db Querier, groupID string, userID string, status string) error {
	_, err := db.Exec(ctx, `
		insert into group_payments (
			group_id,
			user_id,
			status,
			amount_expected
		)
		select
			g.id,
			$2,
			$3,
			case when $3 = 'exempt' then 0 else g.payment_amount end
		from groups g
		where g.id = $1
			and g.is_paid = true
		on conflict (group_id, user_id) do nothing
	`, groupID, userID, status)

	return err
}

func ListGroupPayments(ctx context.Context, db Querier, groupID string) ([]dto.GroupPaymentResponse, error) {
	rows, err := db.Query(ctx, `
		select
			gp.id,
			gp.group_id,
			gp.user_id,
			gm.display_name,
			gp.status,
			gp.amount_expected::float8,
			gp.amount_paid::float8,
			gp.payment_method,
			gp.paid_at,
			gp.marked_by_admin_id::text,
			gp.notes,
			gp.created_at,
			gp.updated_at
		from group_payments gp
		join group_members gm on gm.group_id = gp.group_id
			and gm.user_id = gp.user_id
			and gm.status = 'active'
		where gp.group_id = $1
		order by
			case gp.status
				when 'pending' then 0
				when 'paid' then 1
				when 'exempt' then 2
				else 3
			end,
			gm.display_name asc,
			gm.joined_at asc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := []dto.GroupPaymentResponse{}
	for rows.Next() {
		var payment dto.GroupPaymentResponse
		if err := rows.Scan(
			&payment.ID,
			&payment.GroupID,
			&payment.UserID,
			&payment.DisplayName,
			&payment.Status,
			&payment.AmountExpected,
			&payment.AmountPaid,
			&payment.PaymentMethod,
			&payment.PaidAt,
			&payment.MarkedByAdminID,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, rows.Err()
}

func UpdateGroupPayment(ctx context.Context, db Querier, groupID string, userID string, adminID string, request dto.UpdateGroupPaymentRequest) (dto.GroupPaymentResponse, error) {
	var payment dto.GroupPaymentResponse
	err := db.QueryRow(ctx, `
		update group_payments gp
		set
			status = $4,
			amount_paid = $5,
			payment_method = $6,
			notes = $7,
			paid_at = case
				when $4 = 'paid' then coalesce(gp.paid_at, now())
				else gp.paid_at
			end,
			marked_by_admin_id = $3,
			updated_at = now()
		from group_members gm
		where gp.group_id = $1
			and gp.user_id = $2
			and gm.group_id = gp.group_id
			and gm.user_id = gp.user_id
			and gm.status = 'active'
		returning
			gp.id,
			gp.group_id,
			gp.user_id,
			gm.display_name,
			gp.status,
			gp.amount_expected::float8,
			gp.amount_paid::float8,
			gp.payment_method,
			gp.paid_at,
			gp.marked_by_admin_id::text,
			gp.notes,
			gp.created_at,
			gp.updated_at
	`, groupID, userID, adminID, request.Status, request.AmountPaid, request.PaymentMethod, request.Notes).Scan(
		&payment.ID,
		&payment.GroupID,
		&payment.UserID,
		&payment.DisplayName,
		&payment.Status,
		&payment.AmountExpected,
		&payment.AmountPaid,
		&payment.PaymentMethod,
		&payment.PaidAt,
		&payment.MarkedByAdminID,
		&payment.Notes,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.GroupPaymentResponse{}, ErrNotFound
	}

	return payment, err
}

func GroupPaymentsSummary(ctx context.Context, db Querier, groupID string) (dto.GroupPaymentsSummaryResponse, error) {
	var summary dto.GroupPaymentsSummaryResponse
	err := db.QueryRow(ctx, `
		select
			count(*)::int as total_participants,
			count(*) filter (where status = 'paid')::int as paid_count,
			count(*) filter (where status = 'pending')::int as pending_count,
			count(*) filter (where status = 'exempt')::int as exempt_count,
			count(*) filter (where status = 'refunded')::int as refunded_count,
			coalesce(sum(amount_expected), 0)::float8 as total_expected,
			coalesce(sum(amount_paid), 0)::float8 as total_paid,
			coalesce(sum(
				case
					when status = 'pending' then greatest(amount_expected - amount_paid, 0)
					else 0
				end
			), 0)::float8 as total_pending
		from group_payments
		where group_id = $1
	`, groupID).Scan(
		&summary.TotalParticipants,
		&summary.PaidCount,
		&summary.PendingCount,
		&summary.ExemptCount,
		&summary.RefundedCount,
		&summary.TotalExpected,
		&summary.TotalPaid,
		&summary.TotalPending,
	)

	return summary, err
}
