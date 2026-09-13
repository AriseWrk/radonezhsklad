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
ErrInvalidRole        = errors.New("invalid role")
)

var validRoles = map[string]bool{"admin": true, "manager": true, "warehouse": true, "user": true}

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
if err != nil { return nil, err }
if existing != nil { return nil, ErrUserExists }

hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil { return nil, err }

u := &models.User{
Email: email, PasswordHash: string(hash), FullName: fullName,
LastName: fullName, Role: "user",
}
if err := s.users.Create(ctx, u); err != nil { return nil, err }
return u, nil
}

type CreateUserInput struct {
Email       string
Password    string
LastName    string
FirstName   string
MiddleName  string
Phone       string
Login       string
Description string
Role        string
}

func (s *AuthService) CreateUserWithRole(ctx context.Context, in CreateUserInput) (*models.User, error) {
if !validRoles[in.Role] { return nil, ErrInvalidRole }
existing, err := s.users.FindByEmail(ctx, in.Email)
if err != nil { return nil, err }
if existing != nil { return nil, ErrUserExists }

hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
if err != nil { return nil, err }

fullName := in.LastName + " " + in.FirstName
if in.MiddleName != "" { fullName += " " + in.MiddleName }

u := &models.User{
Email: in.Email, PasswordHash: string(hash), FullName: fullName,
LastName: in.LastName, FirstName: in.FirstName, MiddleName: in.MiddleName,
Role: in.Role,
}
if in.Phone != "" { u.Phone = &in.Phone }
if in.Login != "" { u.Login = &in.Login } else {
prefix := in.Email
for i, c := range in.Email { if c == '@' { prefix = in.Email[:i]; break } }
u.Login = &prefix
}
if in.Description != "" { u.Description = &in.Description }

if err := s.users.Create(ctx, u); err != nil { return nil, err }
return u, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]models.User, error) { return s.users.List(ctx) }

func (s *AuthService) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
if !validRoles[role] { return ErrInvalidRole }
return s.users.UpdateRole(ctx, id, role)
}

func (s *AuthService) UpdateActive(ctx context.Context, id uuid.UUID, active bool) error {
return s.users.UpdateActive(ctx, id, active)
}

func (s *AuthService) UpdateUser(ctx context.Context, id uuid.UUID, in CreateUserInput) (*models.User, error) {
if !validRoles[in.Role] { return nil, ErrInvalidRole }
u, err := s.users.FindByID(ctx, id)
if err != nil { return nil, err }
if u == nil { return nil, ErrUserExists }

u.LastName = in.LastName
u.FirstName = in.FirstName
u.MiddleName = in.MiddleName
u.Role = in.Role
if in.Phone != "" { u.Phone = &in.Phone } else { u.Phone = nil }
if in.Login != "" { u.Login = &in.Login }
if in.Description != "" { u.Description = &in.Description } else { u.Description = nil }

if err := s.users.Update(ctx, u); err != nil { return nil, err }
return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
u, err := s.users.FindByEmail(ctx, email)
if err != nil { return nil, err }
if u == nil || !u.IsActive { return nil, ErrInvalidCredentials }
if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
return nil, ErrInvalidCredentials
}
return s.issueTokens(ctx, u)
}

func (s *AuthService) issueTokens(ctx context.Context, u *models.User) (*TokenPair, error) {
accessExp := time.Now().Add(time.Duration(s.cfg.AccessTTLMin) * time.Minute)
accessClaims := jwt.MapClaims{
"sub": u.ID.String(), "role": u.Role,
"iat": time.Now().Unix(), "exp": accessExp.Unix(),
}
accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
if err != nil { return nil, err }

refreshRaw := make([]byte, 32)
if _, err := rand.Read(refreshRaw); err != nil { return nil, err }
refreshToken := hex.EncodeToString(refreshRaw)
hash := sha256.Sum256([]byte(refreshToken))
refreshExp := time.Now().Add(time.Duration(s.cfg.RefreshTTLDay) * 24 * time.Hour)
if err := s.tokens.Save(ctx, u.ID, hex.EncodeToString(hash[:]), refreshExp); err != nil { return nil, err }

return &TokenPair{
AccessToken: accessToken, RefreshToken: refreshToken,
TokenType: "Bearer", ExpiresIn: s.cfg.AccessTTLMin * 60,
}, nil
}

type AccessClaims struct {
UserID uuid.UUID
Role   string
}

func (s *AuthService) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { return nil, ErrInvalidToken }
return []byte(s.cfg.JWTSecret), nil
})
if err != nil || !token.Valid { return nil, ErrInvalidToken }
claims, ok := token.Claims.(jwt.MapClaims)
if !ok { return nil, ErrInvalidToken }
sub, _ := claims["sub"].(string)
userID, err := uuid.Parse(sub)
if err != nil { return nil, ErrInvalidToken }
role, _ := claims["role"].(string)
return &AccessClaims{UserID: userID, Role: role}, nil
}