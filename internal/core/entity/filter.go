package entity

import (
	"errors"

	"github.com/google/uuid"
)

type FilterInterface interface {
	Get() ([]Filter, error)
	GetById(id string) (*Filter, error)
	GetByTags(tags map[string]string) ([]Filter, error)
	Create(*Filter) (*Filter, error)
	Update(*Filter) (*Filter, error)
}

type Filter struct {
	Tags     map[string]string `json:"tags,omitempty"`
	Account  string            `json:"account" binding:"required"`
	Provider string            `json:"provider" binding:"required"`
}

type FilterRegister struct {
	ID string `json:"id" binding:"required"`
	Filter
}

func (f *Filter) Validate() error {

	var err error = nil
	if f.Provider == "" {
		err = errors.New("the cloud provider cannot be empty")
	}

	if f.Account == "" {
		err = errors.New("the account cannot be empty")
	}

	return err
}

func (f *FilterRegister) Validate() error {

	if err := f.Filter.Validate(); err != nil {
		return err
	}

	if f.ID == "" {
		return errors.New("the filter register cannot be empty ID")
	}
	return nil
}

func (f *FilterRegister) CreateID() error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	f.ID = id.String()

	return nil
}

func (f *FilterRegister) ValidateID() error {
	if err := uuid.Validate(f.ID); err != nil {
		return err
	}
	return nil
}
