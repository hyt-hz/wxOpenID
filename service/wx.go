package service

import "golang.org/x/net/context"

type WXService struct {
}

func NewWXService() *WXService {
	s := WXService{}
	return &s
}

func (s *WXService) ValidateServer(ctx context.Context) {

}
