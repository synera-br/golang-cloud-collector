package entity

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackstageInterface interface {
	TriggerSyncProvider(ctx context.Context, trigger *Trigger) ([]KindReource, error)
	GetAllKinds(ctx context.Context, filter FilterKind) ([]KindReource, error)
}

type CloudProvider int

// Defina constantes para o enum
const (
	NONE CloudProvider = iota
	AZURE
	AWS
	GCP
)

func (c *CloudProvider) GetProvider(name string) CloudProvider {
	if name == "azure" {
		return AZURE
	}
	if name == "aws" {
		return AWS
	}
	if name == "gcp" {
		return GCP
	}
	return NONE
}

func (c *CloudProvider) FilterName(name int) string {
	if name == 1 {
		return "azure"
	}
	if name == 2 {
		return "aws"
	}
	if name == 3 {
		return "gcp"
	}
	return "none"
}

func (c *CloudProvider) GetName() string {
	if *c == 1 {
		return "azure"
	}
	if *c == 2 {
		return "aws"
	}
	if *c == 3 {
		return "gcp"
	}
	return "none"
}

// FilterResource
// Filtro de recursos da cloud por nome e tipo do recurso
// ResourceName nome do recurso.  Exemplo blablabla
// ResourceType tipo do recurso. Example storageAccount
type FilterResource struct {
	ResourceName string `json:"name,omitempty"`
	ResourceType string `json:"type,omitempty"`
}

type FilterTag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type Trigger struct {
	Provider       string         `json:"provider" binding:"required"`
	TargetResource FilterResource `json:"target_resource,omitempty"`
	TargetTags     FilterTag      `json:"target_tag,omitempty"`
}

type Depends struct {
	Kind  string `json:"kind" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type Metadata struct {
	metav1.ObjectMeta
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type KindReource struct {
	Metadata Metadata `json:"metadata" binding:"required"`
	Spec     Resource `json:"spec" binding:"required"`
	Kind     string   `json:"kind" binding:"required"`
}

type Resource struct {
	Type         string   `json:"type" binding:"required"`
	Owner        string   `json:"owner" binding:"required"`
	System       string   `json:"system,omitempty"`
	DependsOn    []string `json:"dependsOn,omitempty"`
	DependencyOf []string `json:"dependencyOf,omitempty"`
}

type FilterKind struct {
	Name      string `json:"name"`
	Kind      string `json:"kind" `
	Namespace string `json:"namespace" `
}

func (k *KindReource) Validate() error {

	if k.Kind == "" {
		k.Kind = "Resource"
	}

	if k.Metadata.Namespace == "" {
		k.Metadata.Namespace = "default"
	}

	if k.Metadata.Name == "" {
		return errors.New("the resource name cannot be empty")
	}

	if k.Spec.Type == "" {
		return errors.New("the resource type cannot be empty")
	}

	if k.Spec.Owner == "" {
		return errors.New("the resource owner cannot be empty")
	}

	return nil
}
