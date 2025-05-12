package tokens

import (
	"time"

	"math/rand"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	refreshTokenLen = 64	
)

type RefreshToken struct {
	Id uuid.UUID
	UserId uuid.UUID
	Token string
	ClientIp string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (t *RefreshToken) HashToken() []byte {
	hash, _ := bcrypt.GenerateFromPassword([]byte(t.Token), bcrypt.DefaultCost)
	return hash
}

func (t *RefreshToken) GenerateToken() {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, refreshTokenLen)
	for i := 0; i < refreshTokenLen; i++ {
		b[i] = allowed[random.Intn(len(allowed))]
	}
	t.Token = string(b)
}
