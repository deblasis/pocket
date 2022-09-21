package modules

import "github.com/pokt-network/pocket/shared/crypto"

// TODO(olshansky): Show an example of `TypicalUsage`
// TODO(drewsky): Add `Create` function; pocket/issues/163
// TODO(drewsky): Do not embed this inside of modules but force it via an implicit cast at compile time
type Module interface {
	IntegratableModule
	InterruptableModule
}

type IntegratableModule interface {
	SetBus(Bus)
	GetBus() Bus
}

type InterruptableModule interface {
	Start() error
	Stop() error
}

type InitializableModule interface {
	GetModuleName() string
	Create(runtime Runtime) (Module, error)
}

type ConfigurableModule interface {
	ValidateConfig(Config) error
}

type GenesisDependentModule interface {
	ValidateGenesis(GenesisState) error
}

type KeyholderModule interface {
	GetPrivateKey(Runtime) (crypto.PrivateKey, error)
}
