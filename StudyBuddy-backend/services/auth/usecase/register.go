package usecase

import (
	"context"
	"studybuddy/backend/services/auth/domain"
)

// RegisterInput for registration.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// RegisterOutput returned after successful registration.
type RegisterOutput struct {
	UserID       string
	Email        string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64 // unix seconds
}

// Register registers a new user and returns tokens.
type Register interface {
	Register(ctx context.Context, in RegisterInput) (RegisterOutput, error)
}

// RegistrationEmbeddingHook triggers async embedding generation after signup.
type RegistrationEmbeddingHook interface {
	OnUserRegistered(userID string)
}

type register struct {
	repo   UserRepository
	hasher PasswordHasher
	jwt    JWTIssuer
	embed  RegistrationEmbeddingHook
}

// NewRegister creates the Register use case.
func NewRegister(repo UserRepository, hasher PasswordHasher, jwt JWTIssuer, embed RegistrationEmbeddingHook) Register {
	return &register{repo: repo, hasher: hasher, jwt: jwt, embed: embed}
}

func (u *register) Register(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	existing, _ := u.repo.GetByEmail(ctx, in.Email)
	if existing != nil {
		return RegisterOutput{}, domain.ErrEmailExists
	}
	hash, err := u.hasher.Hash(in.Password)
	if err != nil {
		return RegisterOutput{}, err
	}
	user := &domain.User{
		Email:        in.Email,
		PasswordHash: hash,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		Role:         domain.RoleStudent,
		IsActive:     true,
	}
	if err := u.repo.Create(ctx, user); err != nil {
		return RegisterOutput{}, err
	}
	if u.embed != nil {
		u.embed.OnUserRegistered(user.ID)
	}
	access, refresh, expAt, err := u.jwt.IssuePair(user.ID, user.Email, user.Role)
	if err != nil {
		return RegisterOutput{}, err
	}
	return RegisterOutput{
		UserID:       user.ID,
		Email:        user.Email,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expAt,
	}, nil
}
