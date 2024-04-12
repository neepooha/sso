package permissions

import (
	"context"
	"errors"
	perm "sso/internal/services/permissions"

	ssov1 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Perm interface {
	SetAdmin(ctx context.Context, email string) (setAdmin bool, err error)
	DelAdmin(ctx context.Context, email string) (delAdmin bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedPermissionsServer
	perm Perm
}

func Register(gRPC *grpc.Server, perm Perm) {
	ssov1.RegisterPermissionsServer(gRPC, &serverAPI{perm: perm})
}

func (s *serverAPI) SetAdmin(ctx context.Context, req *ssov1.SetAdminRequest) (*ssov1.SetAdminResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	setAdmin, err := s.perm.SetAdmin(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.SetAdminResponse{SetAdmin: setAdmin}, nil
}

func (s *serverAPI) DelAdmin(ctx context.Context, req *ssov1.DelAdminRequest) (*ssov1.DelAdminResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	delAdmin, err := s.perm.DelAdmin(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, perm.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.DelAdminResponse{DelAdmin: delAdmin}, nil
}
