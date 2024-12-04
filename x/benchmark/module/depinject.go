package module

import (
	"unsafe"

	"cosmossdk.io/core/telemetry"
	"cosmossdk.io/depinject"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	gen "cosmossdk.io/x/benchmark/generator"
)

const ModuleName = "benchmark"
const maxStoreKeyGenIterations = 100

func init() {
	// TODO try depinject gogo API
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(
			ProvideModule,
		),
	)
}

type StoreKeyRegistrar interface {
	RegisterKey(string)
}

type Input struct {
	depinject.In

	Logger       log.Logger
	Cfg          *modulev1.Module
	Registrar    StoreKeyRegistrar `optional:"true"`
	StoreFactory store.KVStoreServiceFactory
	Telemetry    telemetry.Service
}

func ProvideModule(
	in Input,
) (appmodule.AppModule, error) {
	cfg := in.Cfg
	kvMap := make(KVServiceMap)
	storeKeys, err := gen.StoreKeys(ModuleName, cfg.GenesisParams.Seed, cfg.GenesisParams.BucketCount)
	if err != nil {
		return nil, err
	}
	for _, sk := range storeKeys {
		// app v2 case
		if in.Registrar != nil {
			in.Registrar.RegisterKey(sk)
		}
		kvService := in.StoreFactory(unsafeStrToBytes(sk))
		kvMap[sk] = kvService
	}

	return NewAppModule(cfg.GenesisParams, storeKeys, kvMap, in.Logger, in.Telemetry), nil
}

type KVServiceMap map[string]store.KVStoreService

// unsafeStrToBytes uses unsafe to convert string into byte array. Returned bytes
// must not be altered after this function is called as it will cause a segmentation fault.
func unsafeStrToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s)) // ref https://github.com/golang/go/issues/53003#issuecomment-1140276077
}

// unsafeBytesToStr is meant to make a zero allocation conversion
// from []byte -> string to speed up operations, it is not meant
// to be used generally, but for a specific pattern to delete keys
// from a map.
func unsafeBytesToStr(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
