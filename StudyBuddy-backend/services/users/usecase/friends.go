package usecase

import "context"

// FriendshipStore lists or removes friendships (implemented by the matching repository adapter).
type FriendshipStore interface {
	ListFriends(ctx context.Context, userID string) ([]string, error)
	Delete(ctx context.Context, userID, friendID string) error
}

type ListFriends interface {
	ListFriends(ctx context.Context, userID string) ([]string, error)
}

type listFriends struct {
	store FriendshipStore
}

func NewListFriends(store FriendshipStore) ListFriends {
	return &listFriends{store: store}
}

func (l *listFriends) ListFriends(ctx context.Context, userID string) ([]string, error) {
	return l.store.ListFriends(ctx, userID)
}

type RemoveFriend interface {
	RemoveFriend(ctx context.Context, userID, friendID string) error
}

type removeFriend struct {
	store FriendshipStore
}

func NewRemoveFriend(store FriendshipStore) RemoveFriend {
	return &removeFriend{store: store}
}

func (r *removeFriend) RemoveFriend(ctx context.Context, userID, friendID string) error {
	return r.store.Delete(ctx, userID, friendID)
}
