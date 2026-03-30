package auth

import (
    "errors"
    "christ-api/pkg/jwt"
)

type AuthService struct {
    Repo AuthRepository
}

func (s *AuthService) Login(username, password string) (string, error) {
    user := s.Repo.FindByUsername(username)

    if user == nil {
        return "", errors.New("user not found")
    }

    if user.Password != password {
        return "", errors.New("wrong password")
    }

    token, err := jwt.GenerateToken(user.ID)
    if err != nil {
        return "", err
    }

    return token, nil
}