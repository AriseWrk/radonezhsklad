package service

import (
"context"
"crypto/rand"
"crypto/sha256"
"encoding/hex"
"errors"
"time"

"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"

"github.com/radonezhsklad/auth/internal/config"
"github.com/radonezhsklad/auth/internal/models"
"github.com/radonezhsklad/auth/internal/repository"
)

var (
ErrInvalidCredentials = errors.New("invalid credentials")
ErrUserExists         = errors.New("user already exists")
ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
users  *repository.UserRepo
tokens *repository.TokenRepo
cfg    *config.Config
}

func NewAuthService(users *repository.UserRepo, tokens *repository.TokenRepo, cfg *config.Config) *AuthService {
return &AuthService{users: users, tokens: tokens, cfg: cfg}
}

type TokenPair struct {
AccessToken  string `json:"access_token"`
RefreshToken string `json:"refresh_token"`
TokenType    string `json:"token_type"`
ExpiresIn    int    `json:"expires_in"`
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*models.User, error) {
existing, err := s.users.FindByEmail(ctx, email)
if err != nil {
return nil, err
}
if existing != nil {
return nil, ErrUserExists
}

hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
return nil, err
}

u := &models.User{
Email:        email,
PasswordHash: string(hash),
FullName:     fullName,
Role:         "user",
}
if err := s.users.Create(ctx, u); err != nil {
return nil, err
}
return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
u, err := s.users.FindByEmail(ctx, email)
if err != nil {
return nil, err
}
if u == nil || !u.IsActive {
return nil, ErrInvalidCredentials
}
if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
return nil, ErrInvalidCredentials
}
return s.issueTokens(ctx, u)
}

func (s *AuthService) issueTokens(ctx context.Context, u *models.User) (*TokenPair, error) {
accessExp := time.Now().Add(time.Duration(s.cfg.AccessTTLMin) * time.Minute)

accessClaims := jwt.MapClaims{
"sub":  u.ID.String(),
"role": u.Role,
"iat":  time.Now().Unix(),
"exp":  accessExp.Unix(),
}
accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
if err != nil {
return nil, err
}

refreshRaw := make([]byte, 32)
if _, err := rand.Read(refreshRaw); err != nil {
return nil, err
}
refreshToken := hex.EncodeToString(refreshRaw)
hash := sha256.Sum256([]byte(refreshToken))
refreshExp := time.Now().Add(time.Duration(s.cfg.RefreshTTLDay) * 24 * time.Hour)

if err := s.tokens.Save(ctx, u.ID, hex.EncodeToString(hash[:]), refreshExp); err != nil {
return nil, err
}

return &TokenPair{
AccessToken:  accessToken,
RefreshToken: refreshToken,
TokenType:    "Bearer",
ExpiresIn:    s.cfg.AccessTTLMin * 60,
}, nil
}

type AccessClaims struct {
UserID uuid.UUID
Role   string
}

func (s *AuthService) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, ErrInvalidToken
}
return []byte(s.cfg.JWTSecret), nil
})
if err != nil || !token.Valid {
return nil, ErrInvalidToken
}

claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
return nil, ErrInvalidToken
}

sub, _ := claims["sub"].(string)
userID, err := uuid.Parse(sub)
if err != nil {
return nil, ErrInvalidToken
}
role, _ := claims["role"].(string)

return &AccessClaims{UserID: userID, Role: role}, nil
}