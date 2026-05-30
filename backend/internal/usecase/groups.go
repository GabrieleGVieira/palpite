package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/gabrielevieira/palpitai/backend/internal/apperrors"
	"github.com/gabrielevieira/palpitai/backend/internal/dto"
	"github.com/gabrielevieira/palpitai/backend/internal/repositories"
)

const inviteCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

var (
	ErrGroupFull          = apperrors.NewConflict("group is full")
	ErrGroupNotFound      = apperrors.NewNotFound("group not found")
	ErrGroupOwnerRequired = apperrors.NewForbidden("group owner required")
	ErrPaymentNotFound    = apperrors.NewNotFound("payment not found")
)

type GroupUsecase struct {
	db Datastore
}

func NewGroupUsecase(db Datastore) GroupUsecase {
	return GroupUsecase{db: db}
}

func (uc GroupUsecase) ListGroups(ctx context.Context, userID string) ([]dto.GroupListItemResponse, error) {
	return ListGroups(ctx, uc.db, userID)
}

func (uc GroupUsecase) CreateGroup(ctx context.Context, userID string, displayName string, request dto.CreateGroupRequest) (dto.GroupResponse, error) {
	return CreateGroup(ctx, uc.db, userID, displayName, request)
}

func (uc GroupUsecase) UpdateGroup(ctx context.Context, ownerID string, groupID string, request dto.UpdateGroupRequest) (dto.GroupResponse, error) {
	return UpdateGroup(ctx, uc.db, ownerID, groupID, request)
}

func (uc GroupUsecase) JoinGroup(ctx context.Context, userID string, displayName string, inviteCode string) (dto.JoinGroupResponse, error) {
	return JoinGroup(ctx, uc.db, userID, displayName, inviteCode)
}

func (uc GroupUsecase) ListJoinRequests(ctx context.Context, ownerID string, groupID string) ([]dto.JoinRequestResponse, error) {
	return ListJoinRequests(ctx, uc.db, ownerID, groupID)
}

func (uc GroupUsecase) ListMembers(ctx context.Context, ownerID string, groupID string) ([]dto.GroupMemberResponse, error) {
	return ListMembers(ctx, uc.db, ownerID, groupID)
}

func (uc GroupUsecase) ApproveJoinRequest(ctx context.Context, ownerID string, groupID string, requesterID string) error {
	return ApproveJoinRequest(ctx, uc.db, ownerID, groupID, requesterID)
}

func (uc GroupUsecase) LeaveGroup(ctx context.Context, userID string, groupID string) error {
	return LeaveGroup(ctx, uc.db, userID, groupID)
}

func (uc GroupUsecase) RemoveMember(ctx context.Context, ownerID string, groupID string, memberID string) error {
	return RemoveMember(ctx, uc.db, ownerID, groupID, memberID)
}

func (uc GroupUsecase) TransferOwnership(ctx context.Context, ownerID string, groupID string, nextOwnerID string) error {
	return TransferOwnership(ctx, uc.db, ownerID, groupID, nextOwnerID)
}

func (uc GroupUsecase) ListPayments(ctx context.Context, adminID string, groupID string) ([]dto.GroupPaymentResponse, error) {
	return ListPayments(ctx, uc.db, adminID, groupID)
}

func (uc GroupUsecase) UpdatePayment(ctx context.Context, adminID string, groupID string, userID string, request dto.UpdateGroupPaymentRequest) (dto.GroupPaymentResponse, error) {
	return UpdatePayment(ctx, uc.db, adminID, groupID, userID, request)
}

func (uc GroupUsecase) PaymentsSummary(ctx context.Context, adminID string, groupID string) (dto.GroupPaymentsSummaryResponse, error) {
	return PaymentsSummary(ctx, uc.db, adminID, groupID)
}

func ListGroups(ctx context.Context, db Datastore, userID string) ([]dto.GroupListItemResponse, error) {
	return repositories.ListActiveUserGroups(ctx, db, userID)
}

func JoinGroup(ctx context.Context, db Datastore, userID string, displayName string, inviteCode string) (dto.JoinGroupResponse, error) {
	groupSummary, err := repositories.GroupInviteSummaryByCode(ctx, db, inviteCode)
	if errors.Is(err, repositories.ErrNotFound) {
		return dto.JoinGroupResponse{}, ErrGroupNotFound
	}
	if err != nil {
		return dto.JoinGroupResponse{}, err
	}

	currentStatus, err := repositories.GroupMemberStatus(ctx, db, groupSummary.ID, userID)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return dto.JoinGroupResponse{}, err
	}
	if currentStatus == "pending" {
		group, err := groupByID(ctx, db, groupSummary.ID, userID, "member", "pending")
		return dto.JoinGroupResponse{Group: group, MembershipStatus: "pending"}, err
	}
	if currentStatus == "active" {
		group, err := groupByID(ctx, db, groupSummary.ID, userID, "member", "active")
		return dto.JoinGroupResponse{Group: group, MembershipStatus: "active"}, err
	}

	if groupSummary.ParticipantLimit != nil && groupSummary.MemberCount >= *groupSummary.ParticipantLimit {
		return dto.JoinGroupResponse{}, ErrGroupFull
	}

	nextStatus := "active"
	if groupSummary.IsPrivate {
		nextStatus = "pending"
	}

	err = withTx(ctx, db, func(tx repositories.Querier) error {
		if err := repositories.InsertGroupMember(ctx, tx, groupSummary.ID, userID, nextStatus, displayName); err != nil {
			return err
		}

		if groupSummary.IsPaid && nextStatus == "active" {
			return repositories.InsertPaymentForMemberIfPaidGroup(ctx, tx, groupSummary.ID, userID, string(dto.PaymentStatusPending))
		}

		return nil
	})
	if err != nil {
		return dto.JoinGroupResponse{}, err
	}

	group, err := groupByID(ctx, db, groupSummary.ID, userID, "member", nextStatus)
	if err != nil {
		return dto.JoinGroupResponse{}, err
	}

	return dto.JoinGroupResponse{Group: group, MembershipStatus: nextStatus}, nil
}

func ListJoinRequests(ctx context.Context, db Datastore, ownerID string, groupID string) ([]dto.JoinRequestResponse, error) {
	return repositories.ListPendingJoinRequests(ctx, db, ownerID, groupID)
}

func ListMembers(ctx context.Context, db Datastore, ownerID string, groupID string) ([]dto.GroupMemberResponse, error) {
	return repositories.ListActiveGroupMembers(ctx, db, ownerID, groupID)
}

func ApproveJoinRequest(ctx context.Context, db Datastore, ownerID string, groupID string, requesterID string) error {
	return withTx(ctx, db, func(tx repositories.Querier) error {
		capacity, err := repositories.OwnerGroupCapacity(ctx, tx, ownerID, groupID)
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrGroupNotFound
		}
		if err != nil {
			return err
		}

		if capacity.ParticipantLimit != nil && capacity.MemberCount >= *capacity.ParticipantLimit {
			return ErrGroupFull
		}

		err = repositories.ApprovePendingMember(ctx, tx, groupID, requesterID)
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrGroupNotFound
		}
		if err != nil {
			return err
		}

		return repositories.InsertPaymentForMemberIfPaidGroup(ctx, tx, groupID, requesterID, string(dto.PaymentStatusPending))
	})
}

func LeaveGroup(ctx context.Context, db Datastore, userID string, groupID string) error {
	membership, err := repositories.GroupMembershipByUser(ctx, db, groupID, userID)
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}
	if err != nil {
		return err
	}
	if membership.Role == "owner" {
		return ErrGroupOwnerRequired
	}
	if membership.Status != "active" {
		return ErrGroupNotFound
	}

	err = repositories.DeleteOwnGroupMembership(ctx, db, groupID, userID)
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}

	return err
}

func RemoveMember(ctx context.Context, db Datastore, ownerID string, groupID string, memberID string) error {
	err := repositories.DeleteGroupMemberByOwner(ctx, db, ownerID, groupID, memberID)
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}

	return err
}

func TransferOwnership(ctx context.Context, db Datastore, ownerID string, groupID string, nextOwnerID string) error {
	err := repositories.TransferGroupOwnership(ctx, db, ownerID, groupID, nextOwnerID)
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}

	return err
}

func NormalizeCreateGroupRequest(request dto.CreateGroupRequest) (dto.CreateGroupRequest, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	request.MatchScope = strings.TrimSpace(request.MatchScope)

	if request.Name == "" {
		return request, apperrors.NewValidation("Informe o nome do grupo.")
	}

	if request.MatchScope != "all" && request.MatchScope != "selected" {
		return request, apperrors.NewValidation("Informe uma abrangencia de jogos valida.")
	}

	if request.MatchScope == "all" {
		request.SelectedTeams = []string{}
	}

	if request.MatchScope == "selected" {
		request.SelectedTeams = normalizeTeams(request.SelectedTeams)
		if len(request.SelectedTeams) == 0 {
			return request, apperrors.NewValidation("Selecione pelo menos uma selecao.")
		}
	}

	if request.HasUnlimitedParticipants {
		request.ParticipantLimit = nil
	} else if request.ParticipantLimit == nil || *request.ParticipantLimit < 2 {
		return request, apperrors.NewValidation("O limite precisa ser maior que 1.")
	}

	if err := normalizePaymentFields(&request.IsPaid, &request.PaymentAmount); err != nil {
		return request, err
	}

	return request, nil
}

func NormalizeUpdateGroupRequest(request dto.UpdateGroupRequest) (dto.UpdateGroupRequest, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)

	if request.Name == "" {
		return request, apperrors.NewValidation("Informe o nome do grupo.")
	}

	if request.HasUnlimitedParticipants {
		request.ParticipantLimit = nil
	} else if request.ParticipantLimit == nil || *request.ParticipantLimit < 2 {
		return request, apperrors.NewValidation("O limite precisa ser maior que 1.")
	}

	if err := normalizePaymentFields(&request.IsPaid, &request.PaymentAmount); err != nil {
		return request, err
	}

	return request, nil
}

func NormalizeInviteCode(inviteCode string) string {
	inviteCode = strings.TrimSpace(inviteCode)
	inviteCode = strings.ToUpper(inviteCode)
	inviteCode = strings.ReplaceAll(inviteCode, " ", "")
	inviteCode = strings.ReplaceAll(inviteCode, "-", "")

	return inviteCode
}

func CreateGroup(ctx context.Context, db Datastore, userID string, displayName string, request dto.CreateGroupRequest) (dto.GroupResponse, error) {
	var group dto.GroupResponse
	err := withTx(ctx, db, func(tx repositories.Querier) error {
		for range 5 {
			inviteCode, err := generateInviteCode()
			if err != nil {
				return err
			}

			group, err = repositories.InsertGroupWithOwner(ctx, tx, userID, displayName, request, inviteCode)
			if err == nil {
				if group.IsPaid {
					return repositories.InsertPaymentForMemberIfPaidGroup(ctx, tx, group.ID, userID, string(dto.PaymentStatusExempt))
				}
				return nil
			}

			if repositories.IsUniqueViolation(err) {
				continue
			}

			return err
		}

		return errors.New("failed to generate unique invite code")
	})
	if err != nil {
		return dto.GroupResponse{}, err
	}

	return group, nil
}

func UpdateGroup(ctx context.Context, db Datastore, ownerID string, groupID string, request dto.UpdateGroupRequest) (dto.GroupResponse, error) {
	group, err := repositories.UpdateOwnedGroup(ctx, db, ownerID, groupID, request)
	if errors.Is(err, repositories.ErrNotFound) {
		return dto.GroupResponse{}, ErrGroupNotFound
	}
	if err != nil {
		return dto.GroupResponse{}, err
	}

	return group, nil
}

func ListPayments(ctx context.Context, db Datastore, adminID string, groupID string) ([]dto.GroupPaymentResponse, error) {
	if err := ensureGroupOwner(ctx, db, adminID, groupID); err != nil {
		return nil, err
	}
	if err := repositories.InsertMissingPaymentsForPaidGroup(ctx, db, groupID); err != nil {
		return nil, err
	}

	return repositories.ListGroupPayments(ctx, db, groupID)
}

func UpdatePayment(ctx context.Context, db Datastore, adminID string, groupID string, userID string, request dto.UpdateGroupPaymentRequest) (dto.GroupPaymentResponse, error) {
	request.Status = strings.TrimSpace(request.Status)
	request.PaymentMethod = strings.TrimSpace(request.PaymentMethod)
	request.Notes = strings.TrimSpace(request.Notes)

	if !isValidPaymentStatus(request.Status) {
		return dto.GroupPaymentResponse{}, apperrors.NewValidation("Informe um status de pagamento valido.")
	}
	if request.AmountPaid < 0 {
		return dto.GroupPaymentResponse{}, apperrors.NewValidation("O valor pago não pode ser negativo.")
	}
	if request.AmountExpected < 0 {
		return dto.GroupPaymentResponse{}, apperrors.NewValidation("O valor esperado não pode ser negativo.")
	}

	var payment dto.GroupPaymentResponse
	err := withTx(ctx, db, func(tx repositories.Querier) error {
		if err := ensureGroupOwner(ctx, tx, adminID, groupID); err != nil {
			return err
		}
		if err := repositories.EnsureActiveGroupMember(ctx, tx, groupID, userID); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}
		if err := repositories.InsertMissingPaymentsForPaidGroup(ctx, tx, groupID); err != nil {
			return err
		}

		var updateErr error
		payment, updateErr = repositories.UpdateGroupPayment(ctx, tx, groupID, userID, adminID, request)
		if errors.Is(updateErr, repositories.ErrNotFound) {
			return ErrPaymentNotFound
		}

		return updateErr
	})
	if err != nil {
		return dto.GroupPaymentResponse{}, err
	}

	return payment, nil
}

func PaymentsSummary(ctx context.Context, db Datastore, adminID string, groupID string) (dto.GroupPaymentsSummaryResponse, error) {
	if err := ensureGroupOwner(ctx, db, adminID, groupID); err != nil {
		return dto.GroupPaymentsSummaryResponse{}, err
	}
	if err := repositories.InsertMissingPaymentsForPaidGroup(ctx, db, groupID); err != nil {
		return dto.GroupPaymentsSummaryResponse{}, err
	}

	payments, err := repositories.ListGroupPayments(ctx, db, groupID)
	if err != nil {
		return dto.GroupPaymentsSummaryResponse{}, err
	}

	return CalculatePaymentsSummary(payments), nil
}

func groupByID(ctx context.Context, db Datastore, groupID string, userID string, role string, status string) (dto.GroupListItemResponse, error) {
	group, err := repositories.GroupListItemByID(ctx, db, groupID)
	if errors.Is(err, repositories.ErrNotFound) {
		return dto.GroupListItemResponse{}, ErrGroupNotFound
	}
	if err != nil {
		return dto.GroupListItemResponse{}, err
	}

	group.Role = role
	group.Status = status

	if group.OwnerID == userID {
		group.Role = "owner"
	}

	return group, nil
}

func normalizeTeams(teams []string) []string {
	seen := map[string]bool{}
	normalizedTeams := make([]string, 0, len(teams))

	for _, team := range teams {
		team = strings.TrimSpace(team)
		if team == "" || seen[team] {
			continue
		}

		seen[team] = true
		normalizedTeams = append(normalizedTeams, team)
	}

	return normalizedTeams
}

func normalizePaymentFields(isPaid *bool, amount *float64) error {
	if !*isPaid {
		*amount = 0
		return nil
	}
	if *amount < 0 {
		return apperrors.NewValidation("O valor de participacao não pode ser negativo.")
	}

	return nil
}

func ensureGroupOwner(ctx context.Context, db repositories.Querier, adminID string, groupID string) error {
	err := repositories.EnsureGroupOwner(ctx, db, adminID, groupID)
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrGroupNotFound
	}

	return err
}

func isValidPaymentStatus(status string) bool {
	switch dto.PaymentStatus(status) {
	case dto.PaymentStatusPending, dto.PaymentStatusPaid, dto.PaymentStatusExempt, dto.PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

func CalculatePaymentsSummary(payments []dto.GroupPaymentResponse) dto.GroupPaymentsSummaryResponse {
	var summary dto.GroupPaymentsSummaryResponse
	summary.TotalParticipants = len(payments)

	for _, payment := range payments {
		summary.TotalExpected += payment.AmountExpected
		summary.TotalPaid += payment.AmountPaid

		switch dto.PaymentStatus(payment.Status) {
		case dto.PaymentStatusPaid:
			summary.PaidCount++
		case dto.PaymentStatusPending:
			summary.PendingCount++
			pending := payment.AmountExpected - payment.AmountPaid
			if pending > 0 {
				summary.TotalPending += pending
			}
		case dto.PaymentStatusExempt:
			summary.ExemptCount++
		case dto.PaymentStatusRefunded:
			summary.RefundedCount++
		}
	}

	return summary
}

func generateInviteCode() (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	for index, value := range buffer {
		buffer[index] = inviteCodeAlphabet[int(value)%len(inviteCodeAlphabet)]
	}

	return string(buffer), nil
}
