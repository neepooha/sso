package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/services/auth"
	"strings"

	"github.com/go-playground/validator/v10"
	ssov2 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appName string) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID uint64, err error)
	GetUserID(ctx context.Context, email string) (userID uint64, err error)
}

type serverAPI struct {
	ssov2.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov2.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

type LoginRequest struct {
	Email   string `validate:"required,email"`
	Pass    string `validate:"required"`
	AppName string `validate:"required"`
}

type RegisterRequest struct {
	Email string `validate:"required,email"`
	Pass  string `validate:"required,gt=7"`
}

type GetUserIDRequest struct {
	Email string `validate:"required,email"`
}

func (s *serverAPI) Login(ctx context.Context, req *ssov2.LoginRequest) (*ssov2.LoginResponse, error) {
	if err := ValidateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppName())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov2.RegisterRequest) (*ssov2.RegisterResponse, error) {
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
	return &ssov2.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) GetUserID(ctx context.Context, req *ssov2.GetUserIDRequest) (*ssov2.GetUserIDResponse, error) {
	if err := ValidateGetUserID(req); err != nil {
		return nil, err
	}

	id, err := s.auth.GetUserID(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.GetUserIDResponse{UserId: id}, nil
}

func ValidateLogin(req *ssov2.LoginRequest) error {
	var loginReq LoginRequest
	loginReq.Email = req.GetEmail()
	loginReq.Pass = req.GetPassword()
	loginReq.AppName = req.GetAppName()

	if err := validator.New().Struct(loginReq); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateRegister(req *ssov2.RegisterRequest) error {
	var regiserReq RegisterRequest
	regiserReq.Email = req.GetEmail()
	regiserReq.Pass = req.GetPassword()
	if err := validator.New().Struct(regiserReq); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateGetUserID(req *ssov2.GetUserIDRequest) error {
	var loginReq GetUserIDRequest
	loginReq.Email = req.GetEmail()

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
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid", err.Field()))
		}
	}

	return errors.New(strings.Join(errMsgs, ", "))
}
