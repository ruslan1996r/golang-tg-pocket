package repository

// Bucket - это некая сущность в BoltDB, аналог таблиц в SQL
type Bucket string

const (
	AccessTokens  Bucket = "access_tokens"
	RequestTokens Bucket = "request_tokens"
)

// TokenRepository - В БД будут храниться только репозитории
type TokenRepository interface {
	Save(chatID int64, token string, bucket Bucket) error
	Get(chatID int64, bucket Bucket) (string, error)
}
