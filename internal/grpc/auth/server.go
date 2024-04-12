package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/services/auth"
	"strings"

	"github.com/go-playground/validator/v10"
	ssov1 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID uint64, err error)
	IsAdmin(ctx context.Context, userID uint64) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const emptyValue = 0

type LoginRequest struct {
	Email string `validate:"required,email"`
	Pass  string `validate:"required"`
	AppId int32  `validate:"required"`
}

type RegisterRequest struct {
	Email string `validate:"required,email"`
	Pass  string `validate:"required,len=8"`
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := ValidateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := ValidateRegister(req); err != nil {
		return nil, err
	}
	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.Internal, "user_ID is required")
	}
	isadmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov1.IsAdminResponse{IsAdmin: isadmin}, nil
}

func ValidateLogin(req *ssov1.LoginRequest) error {
	var loginReq LoginRequest
	loginReq.Email = req.GetEmail()
	loginReq.Pass = req.GetPassword()
	loginReq.AppId = req.GetAppId()

	if err := validator.New().Struct(loginReq); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateRegister(req *ssov1.RegisterRequest) error {
	var loginReq LoginRequest
	loginReq.Email = req.GetEmail()
	loginReq.Pass = req.GetPassword()

	if err := validator.New().Struct(loginReq); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidationError(errs validator.ValidationErrors) error {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid email", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return errors.New(strings.Join(errMsgs, ", "))
}
