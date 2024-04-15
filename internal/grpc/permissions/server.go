package permissions

import (
	"context"
	"errors"
	"fmt"
	perm "sso/internal/services/permissions"
	"strings"

	"github.com/go-playground/validator/v10"
	ssov2 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Perm interface {
	SetAdmin(ctx context.Context, email string, appName string) (bool, error)
	DelAdmin(ctx context.Context, email string, appName string) (bool, error)
	IsAdmin(ctx context.Context, userID uint64, appName string) (bool, error)
	IsCreator(ctx context.Context, userID uint64, appName string) (bool, error)
}

type SetDelAdminReq struct {
	Email   string `validate:"required,email"`
	AppName string `validate:"required"`
}

type IsAdmin struct {
	UserID   uint64 `validate:"required"`
	AppName string `validate:"required"`
}

type IsCreator struct {
	UserID   uint64 `validate:"required"`
	AppName string `validate:"required"`
}

type serverAPI struct {
	ssov2.UnimplementedPermissionsServer
	perm Perm
}

func Register(gRPC *grpc.Server, perm Perm) {
	ssov2.RegisterPermissionsServer(gRPC, &serverAPI{perm: perm})
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov2.IsAdminRequest) (*ssov2.IsAdminResponse, error) {
	if err := ValidateIsAdm(req); err != nil {
		return nil, err
	}

	isadmin, err := s.perm.IsAdmin(ctx, req.GetUserId(), req.AppName)
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov2.IsAdminResponse{IsAdmin: isadmin}, nil
}

func (s *serverAPI) IsCreator(ctx context.Context, req *ssov2.IsCreatorRequest) (*ssov2.IsCreatorResponse, error) {
	if err := ValidateIsCreator(req); err != nil {
		return nil, err
	}
	iscreator, err := s.perm.IsCreator(ctx, req.GetUserId(), req.GetAppName())
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov2.IsCreatorResponse{IsCreator: iscreator}, nil
}

func (s *serverAPI) SetAdmin(ctx context.Context, req *ssov2.SetAdminRequest) (*ssov2.SetAdminResponse, error) {
	if err := ValidateSet(req); err != nil {
		return nil, err
	}

	setAdmin, err := s.perm.SetAdmin(ctx, req.GetEmail(), req.GetAppName())
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.SetAdminResponse{SetAdmin: setAdmin}, nil
}

func (s *serverAPI) DelAdmin(ctx context.Context, req *ssov2.DelAdminRequest) (*ssov2.DelAdminResponse, error) {
	if err := ValidateDel(req); err != nil {
		return nil, err
	}

	delAdmin, err := s.perm.DelAdmin(ctx, req.GetEmail(), req.GetAppName())
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.DelAdminResponse{DelAdmin: delAdmin}, nil
}

func ValidateSet(req *ssov2.SetAdminRequest) error {
	var reqStruct SetDelAdminReq
	reqStruct.Email = req.GetEmail()
	reqStruct.AppName = req.GetAppName()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateDel(req *ssov2.DelAdminRequest) error {
	var reqStruct SetDelAdminReq
	reqStruct.Email = req.GetEmail()
	reqStruct.AppName = req.GetAppName()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateIsAdm(req *ssov2.IsAdminRequest) error {
	var reqStruct IsAdmin
	reqStruct.UserID = req.GetUserId()
	reqStruct.AppName = req.GetAppName()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateIsCreator(req *ssov2.IsCreatorRequest) error {
	var reqStruct IsCreator
	reqStruct.UserID = req.GetUserId()
	reqStruct.AppName = req.GetAppName()

	if err := validator.New().Struct(reqStruct); err != nil {
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
