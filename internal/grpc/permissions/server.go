package permissions

import (
	"context"
	"errors"
	"fmt"
	perm "sso/internal/services/permissions"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/go-playground/validator/v10"
	ssov2 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Perm interface {
	SetAdmin(ctx context.Context, email string, appID int) (bool, error)
	DelAdmin(ctx context.Context, email string, appID int) (bool, error)
	IsAdmin(ctx context.Context, userID uint64, appID int) (bool, error)
	IsCreator(ctx context.Context, token string, appID int) (bool, error)
}

type SetDelAdminReq struct {
	Email string `validate:"required,email"`
	AppId int32  `validate:"required"`
}

type IsAdmin struct {
	UserID uint32 `validate:"required"`
	AppId  int32  `validate:"required"`
}

type IsCreator struct {
	Token string `validate:"required"`
	AppId int32  `validate:"required"`
}

type serverAPI struct {
	ssov2.UnimplementedPermissionsServer
	perm Perm
}

func Register(gRPC *grpc.Server, perm Perm) {
	ssov2.RegisterPermissionsServer(gRPC, &serverAPI{perm: perm})
}

const emptyValue = 0

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov2.IsAdminRequest) (*ssov2.IsAdminResponse, error) {
	if err := ValidateIsAdm(req); err != nil {
		return nil, err
	}

	isadmin, err := s.perm.IsAdmin(ctx, req.GetUserId(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, perm.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov2.IsAdminResponse{IsAdmin: isadmin}, nil
}

func (s *serverAPI) IsCreator(ctx context.Context, req *ssov2.IsCreatorRequest) (*ssov2.IsCreatorResponse, error) {
	if err := ValidateIsCreator(req); err != nil {
		return nil, err
	}
	iscreator, err := s.perm.IsCreator(ctx, req.GetToken(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, perm.ErrUserNotFound) {
			return nil, status.Error(codes.InvalidArgument, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &ssov2.IsCreatorResponse{IsCreator: iscreator}, nil
}

func (s *serverAPI) SetAdmin(ctx context.Context, req *ssov2.SetAdminRequest) (*ssov2.SetAdminResponse, error) {
	if err := ValidateSet(req); err != nil {
		return nil, err
	}

	tokenStr, err := exractToken(ctx)
	if err != nil {
		return nil, err
	}
	isCreator, err := s.perm.IsCreator(ctx, tokenStr, int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, err
	}
	if !isCreator {
		return nil, status.Error(codes.PermissionDenied, `u r not creator!`)
	}

	setAdmin, err := s.perm.SetAdmin(ctx, req.GetEmail(), int(req.GetAppId()))
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

	tokenStr, err := exractToken(ctx)
	if err != nil {
		return nil, err
	}
	isCreator, err := s.perm.IsCreator(ctx, tokenStr, int(req.GetAppId()))
	if err != nil {
		return nil, err
	}
	if !isCreator {
		return nil, status.Error(codes.PermissionDenied, `u r not creator!`)
	}

	delAdmin, err := s.perm.DelAdmin(ctx, req.GetEmail(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov2.DelAdminResponse{DelAdmin: delAdmin}, nil
}

func exractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers in request")
	}
	authHeaders, ok := md["authorization"]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no header in request")
	}
	if len(authHeaders) != 1 {
		return "", status.Error(codes.Unauthenticated, "more than 1 header in request")
	}
	auth := authHeaders[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", status.Error(codes.Unauthenticated, `missing "Bearer " prefix in "Authorization" header`)
	}
	if auth[len(prefix):] == "" {
		return "", status.Error(codes.Unauthenticated, `missing token in "Authorization" header`)
	}
	return auth[len(prefix):], nil
}

func ValidateSet(req *ssov2.SetAdminRequest) error {
	var reqStruct SetDelAdminReq
	reqStruct.Email = req.GetEmail()
	reqStruct.AppId = req.GetAppId()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateDel(req *ssov2.DelAdminRequest) error {
	var reqStruct SetDelAdminReq
	reqStruct.Email = req.GetEmail()
	reqStruct.AppId = req.GetAppId()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateIsAdm(req *ssov2.IsAdminRequest) error {
	var reqStruct IsAdmin
	reqStruct.UserID = uint32(req.GetUserId())
	reqStruct.AppId = req.GetAppId()

	if err := validator.New().Struct(reqStruct); err != nil {
		validateErr := err.(validator.ValidationErrors)
		return ValidationError(validateErr)
	}
	return nil
}

func ValidateIsCreator(req *ssov2.IsCreatorRequest) error {
	var reqStruct IsCreator
	reqStruct.Token = req.GetToken()
	reqStruct.AppId = req.GetAppId()

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
