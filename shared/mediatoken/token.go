package mediatoken

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const purposeMediaProxy = "media_proxy"

var (
	ErrInvalidToken = errors.New("invalid media token")
	ErrExpiredToken = errors.New("media token expired")
)

type Claims struct {
	ObjectID string `json:"oid"`
	TeamID   string `json:"tid"`
	Purpose  string `json:"purpose"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Issuer{secret: []byte(strings.TrimSpace(secret)), ttl: ttl}
}

func (i *Issuer) Mint(objectID, teamID uuid.UUID) (string, time.Time, error) {
	if len(i.secret) == 0 {
		return "", time.Time{}, fmt.Errorf("media token secret is not configured")
	}
	now := time.Now().UTC()
	expiresAt := now.Add(i.ttl)
	claims := Claims{
		ObjectID: objectID.String(),
		TeamID:   teamID.String(),
		Purpose:  purposeMediaProxy,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (i *Issuer) Parse(raw string) (objectID, teamID uuid.UUID, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(i.secret) == 0 {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return i.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, uuid.Nil, ErrExpiredToken
		}
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.Purpose != purposeMediaProxy {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now().UTC()) {
		return uuid.Nil, uuid.Nil, ErrExpiredToken
	}
	objectID, err = uuid.Parse(claims.ObjectID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	teamID, err = uuid.Parse(claims.TeamID)
	if err != nil {
		return uuid.Nil, uuid.Nil, ErrInvalidToken
	}
	return objectID, teamID, nil
}

func BuildWebhookURL(publicWebhookBase, token string) string {
	base := strings.TrimRight(strings.TrimSpace(publicWebhookBase), "/")
	return fmt.Sprintf("%s/media/%s", base, strings.TrimSpace(token))
}
