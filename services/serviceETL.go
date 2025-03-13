package services

import "context"

type ETLService struct {
}

func NewEtlService() *ETLService {
	return &ETLService{}
}

func (e *ETLService) Name() string {
	return "ETLService"
}

func (e *ETLService) Start(ctx context.Context) error {
	return nil
}

func (e *ETLService) Stop(ctx context.Context) error {
	return nil
}
