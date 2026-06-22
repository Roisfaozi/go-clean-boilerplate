package usecase

import "context"

type actorContextKey string

const actorUserIDKey actorContextKey = "organization.actor_user_id"
const actorRoleKey actorContextKey = "organization.actor_role"

func WithActorUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, actorUserIDKey, userID)
}

func WithActorRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, actorRoleKey, role)
}

func actorUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(actorUserIDKey).(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}

func actorRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(actorRoleKey).(string)
	if !ok || role == "" {
		return "", false
	}
	return role, true
}
