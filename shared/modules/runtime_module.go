package modules

//go:generate mockgen -source=$GOFILE -destination=./mocks/runtime_module_mock.go -aux_files=github.com/pokt-network/pocket/shared/modules=module.go

type RuntimeMgr interface {
	GetConfig() Config
	GetGenesis() GenesisState
}
