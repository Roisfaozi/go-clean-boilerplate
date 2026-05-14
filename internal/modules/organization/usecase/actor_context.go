package usecase

import "context"

type actorContextKey string

const actorUserIDKey actorContextKey = "organization.actor_user_id"

func WithActorUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, actorUserIDKey, userID)
}

func actorUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(actorUserIDKey).(string)
	if !ok || userID == "" {
		return "", false
	}
	return userID, true
}
