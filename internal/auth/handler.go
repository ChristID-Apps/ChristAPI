package auth

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"christ-api/internal/auth/dto/requests"
	"christ-api/internal/auth/helpers"
	"christ-api/internal/contacts"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/idtoken"
)

type Handler struct {
	service *AuthService
}

func NewHandler(repo *AuthRepository) *Handler {
	return &Handler{
		service: &AuthService{Repo: repo},
	}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	req := new(requests.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid login request", err)
	}

	if err := helpers.ValidateLoginRequest(req); err != nil {
		return response.ErrorDetail(c, 422, "Login validation failed", err)
	}

	token, profile, err := h.service.Login(req.Email, req.Password, req.SiteID)
	if err != nil {
		status, message := loginErrorResponse(err)
		return response.ErrorDetail(c, status, message, err)
	}

	data := LoginUserResponseToLoginDataResponse(profile, token)
	return response.Success(c, "Login berhasil", data)
}

func loginErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrInvalidPassword), errors.Is(err, ErrWrongProvider):
		return 401, err.Error()
	case errors.Is(err, ErrAccountInactive), errors.Is(err, ErrAccountPending):
		return 403, err.Error()
	default:
		return 500, "login service unavailable"
	}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	req := new(requests.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid registration request", err)
	}

	if err := helpers.ValidateRegisterRequest(req); err != nil {
		if validationErr, ok := err.(*helpers.ValidationError); ok {
			return response.Error(c, 422, "Registration validation failed", validationErr.Fields)
		}
		return response.ErrorDetail(c, 422, "Registration validation failed", err)
	}

	_, user, contact, err := h.service.RegisterWithContact(req.FullName, req.Phone, req.Address, req.ContactSiteID, req.Email, req.Password, nil, req.SiteID)
	if err != nil {
		status, message := registerErrorResponse(err)
		return response.ErrorDetail(c, status, message, err)
	}

	resp := struct {
		User    interface{}       `json:"user"`
		Contact *contacts.Contact `json:"contact"`
	}{
		User:    UserToUserDTO(user),
		Contact: contact,
	}

	return response.Created(c, "User registered. Please verify using OTP sent.", resp)
}

func registerErrorResponse(err error) (int, string) {
	if errors.Is(err, ErrUserAlreadyExists) {
		return 409, err.Error()
	}
	return 500, "Registration service unavailable"
}

func (h *Handler) VerifyOTP(c *fiber.Ctx) error {
	req := new(requests.VerifyOTPRequest)
	if err := c.BodyParser(req); err != nil {
		log.Printf("otp verification failed stage=parse error=%v", err)
		return response.ErrorDetail(c, 422, "Invalid OTP request", err)
	}

	if err := helpers.ValidateVerifyOTPRequest(req); err != nil {
		log.Printf("otp verification failed stage=validate email=%s error=%v", maskEmail(req.Email), err)
		return response.ErrorDetail(c, 422, "OTP validation failed", err)
	}

	err := h.service.VerifyOTP(req.Email, req.OTPCode)
	if err != nil {
		log.Printf("otp verification failed stage=service email=%s error=%v", maskEmail(req.Email), err)
		return response.ErrorDetail(c, 400, "OTP verification failed", err)
	}

	log.Printf("otp verification succeeded email=%s", maskEmail(req.Email))
	return response.Success(c, "OTP verified successfully. Your account is now pending admin approval.", nil)
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "[invalid-email]"
	}
	local := parts[0]
	if len(local) == 1 {
		return "*@" + parts[1]
	}
	return local[:1] + "***@" + parts[1]
}

func (h *Handler) LoginGoogle(c *fiber.Ctx) error {
	req := new(requests.GoogleLoginRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid Google login request", err)
	}

	if err := helpers.ValidateGoogleLoginRequest(req); err != nil {
		return response.ErrorDetail(c, 422, "Google login validation failed", err)
	}

	payload, err := idtoken.Validate(context.Background(), req.IDToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		log.Printf("❌ Google token verification failed: %v", err)
		return response.ErrorDetail(c, 401, "Invalid Google token", err)
	}

	email, _ := payload.Claims["email"].(string)
	googleID, _ := payload.Claims["sub"].(string)

	if email == "" || googleID == "" {
		return response.ErrorDetail(c, 401, "Invalid Google token claims", errors.New("Google token did not contain email or subject"))
	}

	log.Printf("Google OAuth verified: %s (sub: %s)", email, googleID)

	token, status, profile, err := h.service.GoogleLoginOrRegister(email, googleID, nil)
	if err != nil {
		return response.ErrorDetail(c, 400, "Google login failed", err)
	}

	if status == "pending_username" {
		return response.Success(c, "Google registration successful. Please choose a username.", fiber.Map{
			"status": "pending_username",
			"user":   profile,
		})
	}

	if status != "approved" {
		return response.Success(c, "Google login successful but status is pending/rejected.", fiber.Map{
			"status": status,
			"user":   profile,
		})
	}

	data := LoginUserResponseToLoginDataResponse(profile, token)
	return response.Success(c, "Google Login berhasil", data)
}

func (h *Handler) SubmitGoogleUsername(c *fiber.Ctx) error {
	req := new(requests.SubmitGoogleUsernameRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid Google profile request", err)
	}

	if err := helpers.ValidateSubmitGoogleUsername(req); err != nil {
		return response.ErrorDetail(c, 422, "Google profile validation failed", err)
	}

	payload, err := idtoken.Validate(context.Background(), req.IDToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		return response.Error(c, 401, "Invalid Google token", nil)
	}
	email, _ := payload.Claims["email"].(string)
	googleID, _ := payload.Claims["sub"].(string)
	if email == "" || googleID == "" {
		return response.Error(c, 401, "Invalid Google token claims", nil)
	}

	err = h.service.SubmitGoogleUsername(email, googleID, req.Username, req.FullName, req.Phone, req.Address, req.SiteID)
	if err != nil {
		return response.ErrorDetail(c, 400, "Username submission failed", err)
	}

	return response.Success(c, "Username updated and contact created. Your account is now pending admin approval.", nil)
}

func (h *Handler) GetPendingApprovals(c *fiber.Ctx) error {
	users, err := h.service.GetPendingApprovals()
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to fetch approvals", err)
	}

	dtos := UsersToUserDTOs(users)
	return response.Success(c, "List pending approvals", dtos)
}

func (h *Handler) ApproveUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid User ID", err)
	}

	err = h.service.ApproveUser(int64(id))
	if err != nil {
		return response.ErrorDetail(c, 400, "Approval failed", err)
	}

	return response.Success(c, "User approved successfully.", nil)
}

func (h *Handler) RejectUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid User ID", err)
	}

	err = h.service.RejectUser(int64(id))
	if err != nil {
		return response.ErrorDetail(c, 400, "Rejection failed", err)
	}

	return response.Success(c, "User rejected successfully.", nil)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return response.Error(c, 400, "Invalid user ID", fiber.Map{"detail": "user_id is missing or has an invalid type in request context"})
	}

	err := h.service.Logout(userID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Logout failed", err)
	}

	return response.Success(c, "Logout berhasil", nil)
}
