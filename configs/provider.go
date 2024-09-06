package configs

import (
	"errors"

	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
)

type Provider struct {
	Azure entity.AzureProvider `json:"azure" mapstructure:"azure"`
}

func (p *Provider) Validate() error {

	var hasProvider error = nil
	if !p.Azure.IsEmpty() {
		if err := p.Azure.Validate(); err != nil {
			hasProvider = err
		}
	}

	if p.Azure.IsEmpty() {
		hasProvider = errors.New("no provider has benn configured")
	}

	return hasProvider
}
