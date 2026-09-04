package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/cloudwego/ppt-agent/pkg/db"
	"github.com/cloudwego/ppt-agent/pkg/logger"
)

const (
	jwtDuration     = 7 * 24 * time.Hour
	codeDuration    = 5 * time.Minute
	codeMaxAttempts = 5
)

var jwtSecret = []byte("pptagent")

type jwtClaims struct {
	UserID  uint   `json:"user_id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// SendCode 生成6位验证码，存储并发送邮件
func SendCode(email string) error {
	email = normalizeEmail(email)
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

// LoginWithCode 验证已注册用户的邮箱验证码并返回 JWT 令牌。
// 注册和登录必须分开，不能再通过登录路径隐式创建无密码账号。
func LoginWithCode(email, code string) (token string, user *db.User, err error) {
	email = normalizeEmail(email)
	if email == "" || code == "" {
		return "", nil, errors.New("邮箱和验证码不能为空")
	}

	u := &db.User{}
	if err := db.DB.Where("email = ?", email).First(u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("该邮箱尚未注册，请先注册账号")
		}
		return "", nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if err := consumeVerificationCode(db.DB, email, code); err != nil {
		return "", nil, err
	}
	token, err = createToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("创建会话失败: %w", err)
	}
	return token, u, nil
}

// Register verifies an email code and creates an account with a compliant
// password. The verification code is consumed only when the account is
// successfully created.
func Register(email, code, password string) (token string, user *db.User, err error) {
	email = normalizeEmail(email)
	if !looksLikeEmail(email) || code == "" {
		return "", nil, errors.New("邮箱和验证码不能为空")
	}
	if err := ValidatePassword(password); err != nil {
		return "", nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("密码加密失败: %w", err)
	}
	u := &db.User{Email: email, Password: string(hash)}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var existing db.User
		if err := tx.Where("email = ?", email).First(&existing).Error; err == nil {
			return errors.New("该邮箱已注册，请直接登录")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("查询用户失败: %w", err)
		}
		if err := consumeVerificationCode(tx, email, code); err != nil {
			return err
		}
		if err := tx.Create(u).Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	token, err = createToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("创建会话失败: %w", err)
	}
	return token, u, nil
}

func consumeVerificationCode(database *gorm.DB, email, code string) error {
	var vc db.VerificationCode
	if err := database.Where("email = ? AND code = ? AND used = false AND expires_at > ?",
		email, code, time.Now()).Order("id DESC").First(&vc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("验证码错误或已过期")
		}
		return fmt.Errorf("验证失败: %w", err)
	}
	if err := database.Model(&vc).Update("used", true).Error; err != nil {
		return fmt.Errorf("更新验证码状态失败: %w", err)
	}
	return nil
}

// GuestLoginEnabled keeps the validation-stage entry point easy to turn off
// without changing routes. It is enabled by default and can be disabled with
// GUEST_LOGIN_ENABLED=false (or 0) when the service moves to a closed beta.
func GuestLoginEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("GUEST_LOGIN_ENABLED")))
	return value == "" || (value != "0" && value != "false" && value != "off")
}

// LoginAsGuest creates an isolated, non-recoverable account for the current
// browser. Guest tasks keep normal owner checks and never share a task list.
func LoginAsGuest(clientIP string) (token string, user *db.User, err error) {
	if !GuestLoginEnabled() {
		return "", nil, errors.New("访客体验暂未开放")
	}
	if db.DB == nil {
		return "", nil, errors.New("认证数据库不可用")
	}
	entropy := make([]byte, 12)
	if _, err := rand.Read(entropy); err != nil {
		return "", nil, fmt.Errorf("创建访客会话失败: %w", err)
	}
	user = &db.User{Email: "guest-" + hex.EncodeToString(entropy) + "@guest.local", GuestIPHash: guestIPFingerprint(clientIP)}
	if err := db.DB.Create(user).Error; err != nil {
		return "", nil, fmt.Errorf("创建访客会话失败: %w", err)
	}
	token, err = createToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("创建访客会话失败: %w", err)
	}
	return token, user, nil
}

// guestIPFingerprint is an HMAC, never the source IP. The optional secret is
// deployment-owned; the fallback keeps validation instances functional while
// still avoiding raw IP persistence.
func guestIPFingerprint(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}
	secret := strings.TrimSpace(os.Getenv("GUEST_IP_HASH_SECRET"))
	if secret == "" {
		secret = "deckform-guest-identity-v1"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ip.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

func IsGuestEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	return strings.HasPrefix(email, "guest-") && strings.HasSuffix(email, "@guest.local")
}

// ── JWT token management ─────────────────────────────────────────────────

func createToken(user *db.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:  user.ID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateSession validates the signed JWT without requiring the database.
// Database-backed permission freshness is enforced separately by adminMiddleware.
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
	if claims.UserID == 0 || claims.Email == "" {
		return nil, errors.New("会话缺少用户信息，请重新登录")
	}
	return &db.User{ID: claims.UserID, Email: claims.Email, IsAdmin: claims.IsAdmin}, nil
}

// LoginWithPassword 使用邮箱和密码登录
func LoginWithPassword(email, password string) (token string, user *db.User, err error) {
	email = normalizeEmail(email)
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
		return "", nil, errors.New("该账号尚未设置密码，请使用验证码登录")
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

// SetPassword 为用户设置或更改密码
func SetPassword(userID int, password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	return db.DB.Model(&db.User{}).Where("id = ?", userID).Update("password", string(hash)).Error
}

// SeedRootUser 如果不存在则创建默认账号，已存在则确保 is_admin=true。
// email/password 为空时使用默认 root/root。
func SeedRootUser(email, password string) {
	if email == "" {
		email = "root"
	}
	if password == "" {
		password = "root"
	}
	var u db.User
	err := db.DB.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在则创建
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		u = db.User{Email: email, Password: string(hash), IsAdmin: true}
		if err := db.DB.Create(&u).Error; err != nil {
			logger.Error("seed_root_user_failed", "email", email, "error", err.Error())
			return
		}
		logger.Info("seed_root_user_created", "email", email)
		return
	}
	if err != nil {
		logger.Error("seed_root_user_query_failed", "email", email, "error", err.Error())
		return
	}
	// 存在但非管理员，升级为管理员
	if !u.IsAdmin {
		db.DB.Model(&u).Update("is_admin", true)
		logger.Info("seed_root_user_upgraded", "email", email)
	}
}

// Logout 对于基于 JWT 的认证是空操作。客户端丢弃令牌
func Logout(token string) error {
	return nil
}

// ── Context keys ──────────────────────────────────────────────────────

type ctxKey string

const userIDKey ctxKey = "user_id"
const usernameKey ctxKey = "email"
const isAdminKey ctxKey = "is_admin"

func WithUser(ctx context.Context, u *db.User) context.Context {
	ctx = context.WithValue(ctx, userIDKey, int(u.ID))
	ctx = context.WithValue(ctx, usernameKey, u.Email)
	ctx = context.WithValue(ctx, isAdminKey, u.IsAdmin)
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

func IsAdminFromContext(ctx context.Context) (bool, bool) {
	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	return isAdmin, ok
}

// ValidateUser 根据用户 ID 查询用户记录
func ValidateUser(userID int) (*db.User, error) {
	if db.DB == nil {
		return nil, errors.New("认证数据库不可用")
	}
	var u db.User
	if err := db.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
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

// ValidatePassword is shared by registration and later password changes.
// Keeping the policy here means clients may provide early feedback, while the
// server remains the source of truth.
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 64 {
		return errors.New("密码需为 8–64 位")
	}
	var hasLower, hasUpper, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		return errors.New("密码需同时包含大写字母、小写字母和数字")
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
