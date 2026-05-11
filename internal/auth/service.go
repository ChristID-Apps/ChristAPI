package auth

import (
	"christ-api/internal/contacts"
	"christ-api/pkg/jwt"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo *AuthRepository
}

func (s *AuthService) Login(email, password string, siteID *int64) (string, *LoginUserResponse, error) {
	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("user not found")
	}

	// compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("wrong password")
	}

	// update last login and optional site
	if err := s.Repo.UpdateLastLoginAndSite(user.ID, nil); err != nil {
		// non-fatal for login response, but return error if DB problem
		return "", nil, err
	}

	// if caller provided siteID, update it as well
	if siteID != nil {
		if err := s.Repo.UpdateLastLoginAndSite(user.ID, siteID); err != nil {
			return "", nil, err
		}
	}

	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", nil, err
	}

	profile, err := s.Repo.GetLoginUserProfile(user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, profile, nil
}

func (s *AuthService) Register(email, password string, roleID, siteID, contactID *int64) (string, *User, error) {
	// check if user already exists
	existing, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if existing != nil {
		return "", nil, errors.New("user already exists")
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	user, err := s.Repo.CreateUser(email, string(hashed), roleID, siteID, contactID)
	if err != nil {
		return "", nil, err
	}

	token, err := jwt.GenerateToken(int(user.ID))
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// RegisterWithContact creates a contact and user in a single transaction and returns a token.
func (s *AuthService) RegisterWithContact(fullName string, phone *string, address *string, contactSiteID *int64, email, password string, roleID, userSiteID *int64) (string, *User, *contacts.Contact, error) {
	// check existing user
	existing, err := s.Repo.FindByEmail(email)
	if err != nil {
		return "", nil, nil, err
	}
	if existing != nil {
		return "", nil, nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, nil, err
	}

	c, u, err := s.Repo.CreateContactAndUser(fullName, phone, address, contactSiteID, email, string(hashed), roleID, userSiteID)
	if err != nil {
		return "", nil, nil, err
	}

	token, err := jwt.GenerateToken(int(u.ID))
	if err != nil {
		return "", nil, nil, err
	}

	return token, u, c, nil
}
