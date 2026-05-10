package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/cloudwego/ppt-agent/pkg/db"
)

const (
	jwtDuration    = 7 * 24 * time.Hour
	codeDuration   = 5 * time.Minute
	codeMaxAttempts = 5
)

var jwtSecret = []byte("pptagent")

type jwtClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SendCode generates a 6-digit code, stores it, and emails it.
func SendCode(email string) error {
	if email == "" || !looksLikeEmail(email) {
		return errors.New("请输入有效的邮箱地址")
	}

	var recentCount int64
	db.DB.Model(&db.VerificationCode{}).
		Where("email = ? AND created_at > ?", email, time.Now().Add(-codeDuration)).
		Count(&recentCount)
	if recentCount >= codeMaxAttempts {
		return errors.New("验证码发送过于频繁，请稍后再试")
	}

	code := generateCode(6)

	vc := &db.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(codeDuration),
	}
	if err := db.DB.Create(vc).Error; err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	if err := SendVerificationCode(email, code); err != nil {
		return fmt.Errorf("邮件发送失败: %w", err)
	}
	return nil
}

// LoginWithCode verifies the code and returns a JWT token.
// If the user doesn't exist, it auto-registers them.
func LoginWithCode(email, code string) (token string, user *db.User, isNew bool, err error) {
	if email == "" || code == "" {
		return "", nil, false, errors.New("邮箱和验证码不能为空")
	}

	var vc db.VerificationCode
	if err := db.DB.Where("email = ? AND code = ? AND used = false AND expires_at > ?",
		email, code, time.Now()).Order("id DESC").First(&vc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, false, errors.New("验证码错误或已过期")
		}
		return "", nil, false, fmt.Errorf("验证失败: %w", err)
	}

	db.DB.Model(&vc).Update("used", true)

	u := &db.User{}
	err = db.DB.Where("email = ?", email).First(u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u = &db.User{Email: email}
		if err := db.DB.Create(u).Error; err != nil {
			return "", nil, false, fmt.Errorf("创建用户失败: %w", err)
		}
		isNew = true
	} else if err != nil {
		return "", nil, false, fmt.Errorf("查询用户失败: %w", err)
	}

	token, err = createToken(u)
	if err != nil {
		return "", nil, false, fmt.Errorf("创建会话失败: %w", err)
	}
	return token, u, isNew, nil
}

// ── JWT token management ─────────────────────────────────────────────────

func createToken(user *db.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateSession validates a JWT token and returns the embedded user.
func ValidateSession(tokenString string) (*db.User, error) {
	if tokenString == "" {
		return nil, errors.New("未提供会话令牌")
	}
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("会话无效或已过期，请重新登录")
	}
	return &db.User{
		ID:    claims.UserID,
		Email: claims.Email,
	}, nil
}

// LoginWithPassword logs in with email + password.
func LoginWithPassword(email, password string) (token string, user *db.User, err error) {
	if email == "" || password == "" {
		return "", nil, errors.New("邮箱和密码不能为空")
	}
	u := &db.User{}
	if err := db.DB.Where("email = ?", email).First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("邮箱或密码错误")
		}
		return "", nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if u.Password == "" {
		return "", nil, errors.New("该账号未设置密码，请使用验证码登录后设置密码")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", nil, errors.New("邮箱或密码错误")
	}
	token, err = createToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("创建会话失败: %w", err)
	}
	return token, u, nil
}

// SetPassword sets or changes the password for a user.
func SetPassword(userID int, password string) error {
	if len(password) < 4 {
		return errors.New("密码长度不能少于 4 个字符")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("password", string(hash)).Error
}

// SeedRootUser creates the default root account with password if it doesn't exist.
func SeedRootUser(email, password string) {
	var count int64
	db.DB.Model(&db.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := &db.User{Email: email, Password: string(hash)}
	if err := db.DB.Create(u).Error; err != nil {
		fmt.Printf("[Auth] 默认 root 用户创建失败: %v\n", err)
		return
	}
	fmt.Printf("[Auth] 默认 root 用户已创建: %s\n", email)
}

// Logout is a no-op for JWT-based auth. The client discards the token.
func Logout(token string) error {
	return nil
}

// ── Context keys ──────────────────────────────────────────────────────

type ctxKey string

const userIDKey ctxKey = "user_id"
const usernameKey ctxKey = "email"

func WithUser(ctx context.Context, u *db.User) context.Context {
	ctx = context.WithValue(ctx, userIDKey, int(u.ID))
	ctx = context.WithValue(ctx, usernameKey, u.Email)
	return ctx
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(usernameKey).(string)
	return name, ok
}

// ── Helpers ────────────────────────────────────────────────────────────

func generateCode(digits int) string {
	code := ""
	for i := 0; i < digits; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code
}

func looksLikeEmail(s string) bool {
	return len(s) > 5 && containsRune(s, '@') && containsRune(s, '.')
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
