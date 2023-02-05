package cluster

import (
	"fmt"
	"time"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/bizfly"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/kubernetes"
)

type ProviderEnum string

const (
	BizflyProvider ProviderEnum = "bizfly"
)

func (p ProviderEnum) ConvertMetadataToTenant(
	metadata interface{},
	timeout time.Duration,
) (kubernetes.Tenant, error) {
	switch p {
	case BizflyProvider:
		return bizfly.NewTenantFromMetadata(metadata, timeout)

	default:
		return nil, fmt.Errorf("Don't support %s", p)
	}
}
