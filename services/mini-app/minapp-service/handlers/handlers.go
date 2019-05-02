package handlers

import (
	"context"

	pb "github.com/fanioc/go-poetryminapp/services/mini-app"
)

// NewService returns a naïve, stateless implementation of Service.
func NewService() pb.MinAppServer {
	return minappService{}
}

type minappService struct{}

// Login implements Service.
func (s minappService) Login(ctx context.Context, in *pb.LoginParams) (*pb.Session, error) {
	var resp pb.Session
	resp = pb.Session{
		// Session:
		// UserId:
		// Errcode:
	}
	return &resp, nil
}

// CheckUserSession implements Service.
func (s minappService) CheckUserSession(ctx context.Context, in *pb.CheckSessionParams) (*pb.CheckSession, error) {
	var resp pb.CheckSession
	resp = pb.CheckSession{
		// UserId:
		// Errcode:
	}
	return &resp, nil
}

// GetUserInfo implements Service.
func (s minappService) GetUserInfo(ctx context.Context, in *pb.GetUserInfoParams) (*pb.UserInfo, error) {
	var resp pb.UserInfo
	resp = pb.UserInfo{
		// NickName:
		// Errcode:
	}
	return &resp, nil
}

// UpdateUserInfo implements Service.
func (s minappService) UpdateUserInfo(ctx context.Context, in *pb.UpdateUserInfoParams) (*pb.ErrCode, error) {
	var resp pb.ErrCode
	resp = pb.ErrCode{
		// Code:
		// Msg:
	}
	return &resp, nil
}

// GetUserConfig implements Service.
func (s minappService) GetUserConfig(ctx context.Context, in *pb.GetUserConfigParams) (*pb.UserCofing, error) {
	var resp pb.UserCofing
	resp = pb.UserCofing{
		// NickName:
		// Errcode:
	}
	return &resp, nil
}

// SetUserConfig implements Service.
func (s minappService) SetUserConfig(ctx context.Context, in *pb.SetUserConfigParams) (*pb.ErrCode, error) {
	var resp pb.ErrCode
	resp = pb.ErrCode{
		// Code:
		// Msg:
	}
	return &resp, nil
}
