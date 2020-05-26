package runvm

import (
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmnil"
)

// PluginName is the name of the NodeConn plugin.
const PluginName = "VM"

var (
	// Plugin is the plugin instance of the database plugin.
	Plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	log    *logger.Logger

	vmDaemon   = daemon.New()
	processors = make(map[string]vm.Processor)
)

func configure(_ *node.Plugin) {
	log = logger.NewLogger(PluginName)
}

func run(_ *node.Plugin) {
	_ = RegisterProcessor("7ZLjqJ7ZnASoP9fJwrJ66HadwHR7JzMY2na1jNxVhKBi")

	err := daemon.BackgroundWorker(PluginName, func(shutdownSignal <-chan struct{}) {
		// globally initialize VM
		go vmDaemon.Run()

		<-shutdownSignal

		vmDaemon.Shutdown()
		log.Infof("shutdown VM...  Done")
	})
	if err != nil {
		log.Errorf("failed to start NodeConn worker")
	}
}

// RegisterProcessor creates and registers processor for program hash
// possibly, locates Wasm program code in IPFS and caches here
func RegisterProcessor(programHash string) error {
	switch programHash {
	case "7ZLjqJ7ZnASoP9fJwrJ66HadwHR7JzMY2na1jNxVhKBi":
		processors["7ZLjqJ7ZnASoP9fJwrJ66HadwHR7JzMY2na1jNxVhKBi"] = vmnil.New()
	default:
		return fmt.Errorf("can't create processor for %s", programHash)
	}
	return nil
}

func getProcessor(programHash string) (vm.Processor, error) {
	ret, ok := processors[programHash]
	if !ok {
		return nil, errors.New("no such processor")
	}
	return ret, nil
}

// RunComputationsAsync runs computations in the background and call function upn finishing it
func RunComputationsAsync(ctx *vm.VMTask) error {
	processor, err := getProcessor(ctx.ProgramHash.String())
	if err != nil {
		return err
	}
	builder, err := vm.NewTxBuilder(vm.TransactionBuilderParams{
		Balances:   ctx.Balances,
		OwnColor:   ctx.Color,
		OwnAddress: ctx.Address,
		RequestIds: sctransaction.TakeRequestIds(ctx.Requests),
	})
	if err != nil {
		return err
	}

	reqids := make([]sctransaction.RequestId, len(ctx.Requests))
	for i := range reqids {
		reqids[i] = *ctx.Requests[i].RequestId()
	}

	bh := vm.BatchHash(reqids, ctx.Timestamp)
	taskName := ctx.Address.String() + "." + bh.String()

	err = vmDaemon.BackgroundWorker(taskName, func(shutdownSignal <-chan struct{}) {
		if err := runVM(ctx, builder, processor); err != nil {
			ctx.Log.Errorf("runVM: %v", err)
		}
	})
	return err
}

// runs batch
func runVM(ctx *vm.VMTask, builder *vm.TransactionBuilder, processor vm.Processor) error {
	vmctx := &vm.VMContext{
		Address:       ctx.Address,
		Color:         ctx.Color,
		TxBuilder:     builder,
		Timestamp:     ctx.Timestamp,
		VariableState: state.NewVariableState(ctx.VariableState),
		Log:           ctx.Log,
	}
	stateUpdates := make([]state.StateUpdate, len(ctx.Requests))
	for i, reqRef := range ctx.Requests {
		vmctx.Request = reqRef
		processor.Run(vmctx)
		stateUpdates[i] = vmctx.StateUpdate
		vmctx.VariableState.ApplyStateUpdate(vmctx.StateUpdate)
	}
	var err error
	ctx.ResultBatch, err = state.NewBatch(stateUpdates, ctx.VariableState.StateIndex()+1)
	if err != nil {
		return err
	}
	err = ctx.VariableState.ApplyBatch(ctx.ResultBatch)
	if err != nil {
		return err
	}
	// create final transaction
	ctx.ResultTransaction = vmctx.TxBuilder.Finalize(ctx.VariableState.StateIndex(), ctx.VariableState.Hash())
	// call back
	ctx.OnFinish()
	return nil
}