package domain

import "github.com/spring-financial-group/peacock/pkg/models"

type FeathersUseCase interface {
	GetFeathersFromFile() (*models.Feathers, error)
	GetFeathersFromBytes(data []byte) (*models.Feathers, error)
	ValidateFeathers(f *models.Feathers) error
}
