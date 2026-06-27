package servicetoken

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

const tokenType = "service_account"

var (
	ErrDisabled       = errors.New("service token is disabled")
	ErrInvalidAccount = errors.New("invalid service account")
)

type Config struct {
	Enabled  bool     `mapstructure:"enabled"`
	Issuer   string   `mapstructure:"issuer"`
	TTLDays  int      `mapstructure:"ttl_days"`
	Secret   string   `mapstructure:"secret"`
	Accounts []string `mapstructure:"accounts"`
}

type Manager struct {
	cfg      Config
	accounts map[string]struct{}
}

type Claims struct {
	TokenType string `json:"token_type"`
	UID       int64  `json:"uid"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

func NewManager() *Manager {
	var cfg Config
	_ = viper.UnmarshalKey("service_token", &cfg)

	if envSecret := os.Getenv("ECMDB_SERVICE_TOKEN_SECRET"); envSecret != "" {
		cfg.Secret = envSecret
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "ecmdb"
	}
	if cfg.TTLDays <= 0 {
		cfg.TTLDays = 1095
	}

	accounts := make(map[string]struct{}, len(cfg.Accounts))
	for _, account := range cfg.Accounts {
		account = strings.TrimSpace(account)
		if account == "" {
			continue
		}
		accounts[account] = struct{}{}
	}

	return &Manager{
		cfg:      cfg,
		accounts: accounts,
	}
}

func (m *Manager) Enabled() bool {
	return m != nil && m.cfg.Enabled
}

func (m *Manager) IsServiceAccount(username string) bool {
	if !m.Enabled() {
		return false
	}
	_, ok := m.accounts[username]
	return ok
}

func (m *Manager) Issue(uid int64, username string, ttlDays int) (string, time.Time, error) {
	if !m.Enabled() {
		return "", time.Time{}, ErrDisabled
	}
	if !m.IsServiceAccount(username) {
		return "", time.Time{}, ErrInvalidAccount
	}
	if strings.TrimSpace(m.cfg.Secret) == "" {
		return "", time.Time{}, errors.New("service token secret is empty")
	}
	if ttlDays <= 0 {
		ttlDays = m.cfg.TTLDays
	}
	if ttlDays > m.cfg.TTLDays {
		return "", time.Time{}, fmt.Errorf("ttl_days exceeds configured maximum %d", m.cfg.TTLDays)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(ttlDays) * 24 * time.Hour)
	claims := Claims{
		TokenType: tokenType,
		UID:       uid,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.cfg.Secret))
	return tokenString, expiresAt, err
}

func (m *Manager) Verify(tokenString string) (Claims, error) {
	if !m.Enabled() {
		return Claims{}, ErrDisabled
	}
	if strings.TrimSpace(m.cfg.Secret) == "" {
		return Claims{}, errors.New("service token secret is empty")
	}

	claims := Claims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.cfg.Secret), nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, errors.New("invalid service token")
	}
	if claims.TokenType != tokenType {
		return Claims{}, errors.New("invalid service token type")
	}
	if claims.Issuer != m.cfg.Issuer {
		return Claims{}, errors.New("invalid service token issuer")
	}
	if !m.IsServiceAccount(claims.Username) {
		return Claims{}, ErrInvalidAccount
	}
	if claims.UID == 0 || claims.Username == "" {
		return Claims{}, errors.New("invalid service token claims")
	}
	return claims, nil
}

func ExtractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	fields := strings.Fields(header)
	if len(fields) == 1 {
		return fields[0]
	}
	if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
		return fields[1]
	}
	if len(fields) == 2 && strings.EqualFold(fields[0], "ServiceToken") {
		return fields[1]
	}
	return ""
}
