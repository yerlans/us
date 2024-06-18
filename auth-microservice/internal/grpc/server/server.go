package server

import (
	"auth/internal/domain/models"
	"auth/internal/storage"
	"context"
	"errors"
	pb "github.com/yerlans/us-protos/gen/auth-service"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

type AuthService interface {
	SaveUser(ctx context.Context, email string, pass string) (uid int64, err error)
	GetUser(ctx context.Context, email string) (models.User, error)
	GenerateJWT(user models.User) (string, error)
	ValidateJWT(token string) (models.User, error)
}

type serverAPI struct {
	pb.UnimplementedAuthServiceServer
	authService AuthService
}

func Register(gRPCServer *grpc.Server, authService AuthService) {
	pb.RegisterAuthServiceServer(gRPCServer, &serverAPI{authService: authService})
}

func (s *serverAPI) Register(
	ctx context.Context,
	in *pb.RegisterRequest,
) (*pb.RegisterResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Save the user
	userID, err := s.authService.SaveUser(ctx, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to save user")
	}

	return &pb.RegisterResponse{UserId: int32(userID)}, nil
}

func (s *serverAPI) Login(
	ctx context.Context,
	in *pb.LoginRequest,
) (*pb.LoginResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	user, err := s.authService.GetUser(ctx, in.Email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(in.Password)); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid credentials")
	}

	// Generate JWT
	token, err := s.authService.GenerateJWT(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *serverAPI) ValidateToken(
	ctx context.Context,
	in *pb.ValidateTokenRequest,
) (*pb.ValidateTokenResponse, error) {
	if in.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := s.authService.ValidateJWT(in.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	//TODO: edit proto userId- int
	return &pb.ValidateTokenResponse{
		Email:  user.Email,
		UserId: strconv.FormatInt(user.ID, 10),
	}, nil
}
