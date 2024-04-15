package permissions

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/services/apps"
	"strings"

	"github.com/go-playground/validator/v10"
	ssov2 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Apps interface {
	GetAppID(ctx context.Context, appName string) (int, string, error)
	SetApp(ctx context.Context, email string, appName string, appSecret string) (int, error)
	UpdApp(ctx context.Context, appName string, newAppName string, newAppSecret string) (bool, error)
	DelApp(ctx context.Context, appName string) (bool, error)
}

type GetAppIDReq struct {
	AppName string `validate:"required"`
}

type SetAppReq struct {
	Email     string `validate:"required,email"`
	AppName   string `validate:"required"`
	AppSecret string `validate:"required"`
}

type UpdAppReq struct {
	AppName      string `validate:"required"`
	NewAppName   string `validate:"required"`
	NewAppSecret string `validate:"required"`
}

type DelAppReq struct {
	AppName string `validate:"required"`
}

type serverAPI struct {
	ssov2.UnimplementedAppsServer
	apps Apps
}

func Register(gRPC *grpc.Server, apps Apps) {
	ssov2.RegisterAppsServer(gRPC, &serverAPI{apps: apps})
}

func (s *serverAPI) GetAppID(ctx context.Context, req *ssov2.GetAppRequest) (*ssov2.GetAppResponse, error) {
	if err := ValidateGet(req); err != nil {
		return nil, err
	}

	appID, appName, err := s.apps.GetAppID(ctx, req.GetAppName())
	if err != nil {
		if errors.Is(err, apps.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.GetAppResponse{AppId: int32(appID), AppName: appName}, nil
}

func (s *serverAPI) SetApp(ctx context.Context, req *ssov2.SetAppRequest) (*ssov2.SetAppResponse, error) {
	if err := ValidateSet(req); err != nil {
		return nil, err
	}

	appID, err := s.apps.SetApp(ctx, req.GetEmail(), req.GetAppName(), req.GetAppSecret())
	if err != nil {
		if errors.Is(err, apps.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, apps.ErrAppExists) {
			return nil, status.Error(codes.InvalidArgument, "An application with the same name exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.SetAppResponse{AppID: int32(appID)}, nil
}

func (s *serverAPI) UpdApp(ctx context.Context, req *ssov2.UpdAppRequest) (*ssov2.UpdAppResponse, error) {
	if err := ValidateUpd(req); err != nil {
		return nil, err
	}
	isUpdApp, err := s.apps.UpdApp(ctx, req.GetAppName(), req.GetNewAppName(), req.GetNewAppSecret())
	if err != nil {
		if errors.Is(err, apps.ErrUserNotCreator) {
			return nil, status.Error(codes.Unauthenticated, "You are not creator")
		}
		if errors.Is(err, apps.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, apps.ErrAppExists) {
			return nil, status.Error(codes.InvalidArgument, "an application with the same name exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov2.UpdAppResponse{IsUpdApp: isUpdApp}, nil
}

func (s *serverAPI) DelApp(ctx context.Context, req *ssov2.DelAppRequest) (*ssov2.DelAppResponse, error) {
	if err := ValidateDel(req); err != nil {
		return nil, err
	}

	isDelApp, err := s.apps.DelApp(ctx, req.GetAppName())
	if err != nil {
		if errors.Is(err, apps.ErrUserNotCreator) {
			return nil, status.Error(codes.Unauthenticated, "You are not creator")
		}
		if errors.Is(err, apps.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.DelAppResponse{IsDelApp: isDelApp}, nil
}

func ValidateGet(req *ssov2.GetAppRequest) error {
	var reqStruct GetAppIDReq
	reqStruct.AppName = req.GetAppName()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateSet(req *ssov2.SetAppRequest) error {
	var reqStruct SetAppReq
	reqStruct.Email = req.GetEmail()
	reqStruct.AppName = req.GetAppName()
	reqStruct.AppSecret = req.GetAppSecret()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateUpd(req *ssov2.UpdAppRequest) error {
	var reqStruct UpdAppReq
	reqStruct.AppName = req.GetAppName()
	reqStruct.NewAppName = req.GetNewAppName()
	reqStruct.NewAppSecret = req.GetNewAppSecret()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateDel(req *ssov2.DelAppRequest) error {
	var reqStruct DelAppReq
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
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid", err.Field()))
		}
	}

	return errors.New(strings.Join(errMsgs, ", "))
}
