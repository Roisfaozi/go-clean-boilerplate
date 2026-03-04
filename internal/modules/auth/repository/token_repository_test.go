package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getSessionKey(userID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}

func getAttemptsKey(username string) string {
	return fmt.Sprintf("auth:attempts:%s", username)
}

func getLockedKey(username string) string {
	return fmt.Sprintf("auth:locked:%s", username)
}

func setupGormDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)
	err = db.AutoMigrate(&entity.PasswordResetToken{})
	assert.NoError(t, err)
	return db
}

func TestTokenRepository_StoreToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	now := time.Now().Round(time.Second)
	mockClock := &util.MockClock{CurrentTime: now}
	repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)

	authData := &model.Auth{
		ID:           "session123",
		UserID:       "user456",
		RefreshToken: "some_refresh_token",
		ExpiresAt:    now.Add(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	key := getSessionKey(authData.UserID, authData.ID)

	val, err := json.Marshal(authData)
	assert.NoError(t, err)

	// mockClock.Now() is 'now', authData.ExpiresAt is 'now + 1h'. Diff is exactly 1h.
	mock.ExpectSet(key, val, time.Hour).SetVal("OK")

	err = repo.StoreToken(context.Background(), authData)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_StoreToken_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	now := time.Now().Round(time.Second)
	mockClock := &util.MockClock{CurrentTime: now}
	repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)

	authData := &model.Auth{
		ID:           "session123",
		UserID:       "user456",
		RefreshToken: "some_refresh_token",
		ExpiresAt:    now.Add(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	key := getSessionKey(authData.UserID, authData.ID)
	redisErr := errors.New("redis connection failed")

	val, _ := json.Marshal(authData)
	mock.ExpectSet(key, val, time.Hour).SetErr(redisErr)

	err := repo.StoreToken(context.Background(), authData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), redisErr.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_Save(t *testing.T) {
	db := setupGormDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

	token := &entity.PasswordResetToken{
		Email:     "test@example.com",
		Token:     "token123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := repo.Save(context.Background(), token)
	assert.NoError(t, err)

	var stored entity.PasswordResetToken
	err = db.First(&stored, "email = ?", token.Email).Error
	assert.NoError(t, err)
	assert.Equal(t, token.Token, stored.Token)
}

func TestTokenRepository_FindByToken(t *testing.T) {
	db := setupGormDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

	token := &entity.PasswordResetToken{
		Email:     "test@example.com",
		Token:     "token123",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	db.Create(token)

	result, err := repo.FindByToken(context.Background(), "token123")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, token.Email, result.Email)

	result, err = repo.FindByToken(context.Background(), "invalid")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestTokenRepository_DeleteByEmail(t *testing.T) {
	db := setupGormDB(t)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

	token := &entity.PasswordResetToken{
		Email: "test@example.com",
		Token: "token123",
	}
	db.Create(token)

	err := repo.DeleteByEmail(context.Background(), "test@example.com")
	assert.NoError(t, err)

	var stored entity.PasswordResetToken
	err = db.First(&stored, "email = ?", "test@example.com").Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTokenRepository_GetUserSessions(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	userID := "user123"
	pattern := getSessionKey(userID, "*")

	mock.ExpectKeys(pattern).SetErr(errors.New("redis error"))
	sessions, err := repo.GetUserSessions(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, sessions)
	assert.NoError(t, mock.ExpectationsWereMet())

	keys := []string{getSessionKey(userID, "s1"), getSessionKey(userID, "s2")}
	mock.ExpectKeys(pattern).SetVal(keys)

	s1 := model.Auth{ID: "s1", UserID: userID}
	s2 := model.Auth{ID: "s2", UserID: userID}
	json1, _ := json.Marshal(s1)
	json2, _ := json.Marshal(s2)

	mock.ExpectGet(keys[0]).SetVal(string(json1))
	mock.ExpectGet(keys[1]).SetVal(string(json2))

	sessions, err = repo.GetUserSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "s1", sessions[0].ID)
	assert.Equal(t, "s2", sessions[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectKeys(pattern).SetVal(keys)
	mock.ExpectGet(keys[0]).SetErr(errors.New("get error"))
	mock.ExpectGet(keys[1]).SetVal(string(json2))

	sessions, err = repo.GetUserSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "s2", sessions[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_RevokeAllSessions(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	userID := "user123"
	pattern := getSessionKey(userID, "*")

	mock.ExpectKeys(pattern).SetErr(errors.New("redis error"))
	err := repo.RevokeAllSessions(context.Background(), userID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectKeys(pattern).SetVal([]string{})
	err = repo.RevokeAllSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	keys := []string{"k1", "k2"}
	mock.ExpectKeys(pattern).SetVal(keys)
	mock.ExpectDel(keys...).SetVal(2)
	err = repo.RevokeAllSessions(context.Background(), userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	mock.ExpectKeys(pattern).SetVal(keys)
	mock.ExpectDel(keys...).SetErr(errors.New("del error"))
	err = repo.RevokeAllSessions(context.Background(), userID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})

	userID := "user456"
	sessionID := "session123"
	key := getSessionKey(userID, sessionID)

	expectedAuth := model.Auth{
		ID:           sessionID,
		UserID:       userID,
		RefreshToken: "expected_refresh_token",
	}
	jsonVal, _ := json.Marshal(expectedAuth)

	mock.ExpectGet(key).SetVal(string(jsonVal))

	resultToken, err := repo.GetToken(context.Background(), userID, sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, resultToken)
	assert.Equal(t, expectedAuth.RefreshToken, resultToken.RefreshToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetToken_NotFound(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})

	userID := "user456"
	sessionID := "nonexistent_session"
	key := getSessionKey(userID, sessionID)

	mock.ExpectGet(key).SetErr(redis.Nil)

	resultToken, err := repo.GetToken(context.Background(), userID, sessionID)
	assert.NoError(t, err)
	assert.Nil(t, resultToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetToken_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})

	userID := "user456"
	sessionID := "session123"
	key := getSessionKey(userID, sessionID)
	redisErr := errors.New("redis connection failed")

	mock.ExpectGet(key).SetErr(redisErr)

	resultToken, err := repo.GetToken(context.Background(), userID, sessionID)
	assert.Error(t, err)
	assert.Nil(t, resultToken)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_DeleteToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})

	userID := "user456"
	sessionID := "session123"
	key := getSessionKey(userID, sessionID)

	mock.ExpectDel(key).SetVal(1)

	err := repo.DeleteToken(context.Background(), userID, sessionID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_DeleteToken_RedisError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})

	userID := "user456"
	sessionID := "session123"
	key := getSessionKey(userID, sessionID)
	redisErr := errors.New("redis connection failed")

	mock.ExpectDel(key).SetErr(redisErr)

	err := repo.DeleteToken(context.Background(), userID, sessionID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, redisErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Account Lockout Tests

func TestTokenRepository_GetLoginAttempts(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	username := "testuser"
	key := getAttemptsKey(username)

	mock.ExpectGet(key).SetVal("3")
	attempts, err := repo.GetLoginAttempts(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)

	mock.ExpectGet(key).SetErr(redis.Nil)
	attempts, err = repo.GetLoginAttempts(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, 0, attempts)

	mock.ExpectGet(key).SetErr(errors.New("redis error"))
	attempts, err = repo.GetLoginAttempts(context.Background(), username)
	assert.Error(t, err)
	assert.Equal(t, 0, attempts)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_IncrementLoginAttempts(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	username := "testuser"
	key := getAttemptsKey(username)

	mock.ExpectIncr(key).SetVal(1)
	mock.ExpectExpire(key, time.Hour).SetVal(true)

	attempts, err := repo.IncrementLoginAttempts(context.Background(), username)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)

	mock.ExpectIncr(key).SetErr(errors.New("redis error"))
	// mock.ExpectExpire(key, time.Hour).SetVal(true)
	attempts, err = repo.IncrementLoginAttempts(context.Background(), username)
	assert.Error(t, err)
	assert.Equal(t, 0, attempts)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_ResetLoginAttempts(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	username := "testuser"
	key := getAttemptsKey(username)

	mock.ExpectDel(key).SetVal(1)
	err := repo.ResetLoginAttempts(context.Background(), username)
	assert.NoError(t, err)

	mock.ExpectDel(key).SetErr(errors.New("redis error"))
	err = repo.ResetLoginAttempts(context.Background(), username)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_LockAccount(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	username := "testuser"
	key := getLockedKey(username)
	duration := 30 * time.Minute

	mock.ExpectSet(key, "locked", duration).SetVal("OK")
	err := repo.LockAccount(context.Background(), username, duration)
	assert.NoError(t, err)

	mock.ExpectSet(key, "locked", duration).SetErr(errors.New("redis error"))
	err = repo.LockAccount(context.Background(), username, duration)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_IsAccountLocked(t *testing.T) {
	db, mock := redismock.NewClientMock()
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	repo := repository.NewTokenRepositoryRedis(db, logger, nil, &util.RealClock{})
	username := "testuser"
	key := getLockedKey(username)

	// Locked case
	mock.ExpectTTL(key).SetVal(time.Minute)
	locked, ttl, err := repo.IsAccountLocked(context.Background(), username)
	assert.NoError(t, err)
	assert.True(t, locked)
	assert.Equal(t, time.Minute, ttl)

	// Not locked (key not found)
	mock.ExpectTTL(key).SetVal(-2 * time.Second)
	locked, ttl, err = repo.IsAccountLocked(context.Background(), username)
	assert.NoError(t, err)
	assert.False(t, locked)
	assert.Equal(t, time.Duration(0), ttl) // Assert ttl

	// Redis error
	mock.ExpectTTL(key).SetErr(errors.New("redis error"))
	locked, ttl, err = repo.IsAccountLocked(context.Background(), username)
	assert.Error(t, err)
	assert.False(t, locked)
	assert.Equal(t, time.Duration(0), ttl) // Assert ttl

	assert.NoError(t, mock.ExpectationsWereMet())
}

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) { return 0, nil }
func (w *NoOpWriter) Levels() []logrus.Level    { return logrus.AllLevels }
