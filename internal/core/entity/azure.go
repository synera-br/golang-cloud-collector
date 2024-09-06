package entity

import "errors"

type AzureProviderInterface interface {
	Connection(*AzureProvider) error
}

type AzureProvider struct {
	Subscription      string `json:"subsription" binding:"required" mapstructure:"subsription"`
	ApplicationID     string `json:"application_id" binding:"required" mapstructure:"application_id"`
	ApplicationSecret string `json:"application_sceret" binding:"required" mapstructure:"application_sceret"`
	Tenant            string `json:"tenant" binding:"required" mapstructure:"tenant"`
}

func (a *AzureProvider) Validate() error {

	var err error = nil
	if a.Subscription == "" {
		err = errors.New("the Azure subscription cannot be empty")
	}

	if a.ApplicationID == "" {
		err = errors.New("the Azure application ID cannot be empty")
	}

	if a.ApplicationSecret == "" {
		err = errors.New("the Azure application secret cannot be empty")
	}

	if a.Tenant == "" {
		err = errors.New("the Azure tenant cannot be empty")
	}

	return err
}

func (a *AzureProvider) IsEmpty() bool {
	return a.Subscription == "" &&
		a.ApplicationID == "" &&
		a.ApplicationSecret == "" &&
		a.Tenant == ""
}
