package usecase

import (
	"context"

	"studybuddy/backend/services/users/domain"
)

// AdminListUsersOutput is the paginated result of an admin user list query.
type AdminListUsersOutput struct {
	Users    []domain.UserSummary
	Total    int
	Page     int
	PageSize int
}

// AdminListUsers lists users for the admin panel.
type AdminListUsers interface {
	Execute(ctx context.Context, filter domain.AdminUserFilter, page, pageSize int) (AdminListUsersOutput, error)
}

type adminListUsers struct {
	repo AdminRepository
}

// NewAdminListUsers creates the AdminListUsers use case.
func NewAdminListUsers(repo AdminRepository) AdminListUsers {
	return &adminListUsers{repo: repo}
}

// Execute returns a page of users and pagination metadata.
func (uc *adminListUsers) Execute(ctx context.Context, filter domain.AdminUserFilter, page, pageSize int) (AdminListUsersOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	users, total, err := uc.repo.ListUsers(ctx, filter, pageSize, offset)
	if err != nil {
		return AdminListUsersOutput{}, err
	}
	return AdminListUsersOutput{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
