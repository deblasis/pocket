package cli

import (
	"log"
	"os"
	"sync"

	"github.com/manifoldco/promptui"
	"github.com/pokt-network/pocket/consensus"
	"github.com/pokt-network/pocket/p2p"
	"github.com/pokt-network/pocket/runtime"
	"github.com/pokt-network/pocket/shared"
	pocketCrypto "github.com/pokt-network/pocket/shared/crypto"
	"github.com/pokt-network/pocket/shared/debug"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/pokt-network/pocket/telemetry"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	PromptResetToGenesis         string = "ResetToGenesis"
	PromptPrintNodeState         string = "PrintNodeState"
	PromptTriggerNextView        string = "TriggerNextView"
	PromptTogglePacemakerMode    string = "TogglePacemakerMode"
	PromptShowLatestBlockInStore string = "ShowLatestBlockInStore"

	defaultConfigPath  = "build/config/config1.json"
	defaultGenesisPath = "build/config/genesis.json"
)

var (
	// A P2P module is initialized in order to broadcast a message to the local network
	p2pMod modules.P2PModule

	// A consensus module is initialized in order to get a list of the validator network
	consensusMod modules.ConsensusModule
	modInitOnce  sync.Once

	items = []string{
		PromptResetToGenesis,
		PromptPrintNodeState,
		PromptTriggerNextView,
		PromptTogglePacemakerMode,
		PromptShowLatestBlockInStore,
	}
)

func init() {
	rootCmd.AddCommand(NewDebug())
}

func NewDebug() *cobra.Command {
	c := &cobra.Command{
		Use:   "debug",
		Short: "Debug utility for rapid development",
		Args:  cobra.ExactArgs(0),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initDebug(remoteCLIURL)
		},
		RunE: runDebug,
	}

	return c
}

func runDebug(cmd *cobra.Command, args []string) error {
	for {
		selection, err := promptGetInput()
		if err == nil {
			handleSelect(selection)
		}
	}
}

func promptGetInput() (string, error) {
	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
		Size:  len(items),
	}

	_, result, err := prompt.Run()

	if err == promptui.ErrInterrupt {
		os.Exit(0)
	}

	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func handleSelect(selection string) {
	switch selection {
	case PromptResetToGenesis:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_RESET_TO_GENESIS,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptPrintNodeState:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_PRINT_NODE_STATE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTriggerNextView:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_TRIGGER_NEXT_VIEW,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptTogglePacemakerMode:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_CONSENSUS_TOGGLE_PACE_MAKER_MODE,
			Message: nil,
		}
		broadcastDebugMessage(m)
	case PromptShowLatestBlockInStore:
		m := &debug.DebugMessage{
			Action:  debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE,
			Message: nil,
		}
		sendDebugMessage(m)
	default:
		log.Println("Selection not yet implemented...", selection)
	}
}

// Broadcast to the entire validator set
func broadcastDebugMessage(debugMsg *debug.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	// TODO(olshansky): Once we implement the cleanup layer in RainTree, we'll be able to use
	// broadcast. The reason it cannot be done right now is because this client is not in the
	// address book of the actual validator nodes, so `node1.consensus` never receives the message.
	// p2pMod.Broadcast(anyProto, debug.PocketTopic_DEBUG_TOPIC)

	for _, val := range consensusMod.ValidatorMap() {
		addr, err := pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		p2pMod.Send(addr, anyProto, debug.PocketTopic_DEBUG_TOPIC)
	}
}

// Send to just a single (i.e. first) validator in the set
func sendDebugMessage(debugMsg *debug.DebugMessage) {
	anyProto, err := anypb.New(debugMsg)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create Any proto: %v", err)
	}

	var validatorAddress []byte
	for _, val := range consensusMod.ValidatorMap() {
		validatorAddress, err = pocketCrypto.NewAddress(val.GetAddress())
		if err != nil {
			log.Fatalf("[ERROR] Failed to convert validator address into pocketCrypto.Address: %v", err)
		}
		break
	}

	p2pMod.Send(validatorAddress, anyProto, debug.PocketTopic_DEBUG_TOPIC)
}

func initDebug(remoteCLIURL string) {
	modInitOnce.Do(func() {
		// HACK: rain tree will detect if trying to send to addr=self and not send it
		var err error
		runtimeMgr := runtime.NewManagerFromFiles(defaultConfigPath, defaultGenesisPath, runtime.WithRandomPK())

		consM, err := consensus.Create(runtimeMgr)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create consensus module: %v", err.Error())
		}
		consensusMod = consM.(modules.ConsensusModule)

		p2pM, err := p2p.Create(runtimeMgr)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
		}
		p2pMod = p2pM.(modules.P2PModule)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create p2p module: %v", err.Error())
		}

		// This telemetry module instance is a NOOP because the 'enable_telemetry' flag in the `cfg` above is set to false.
		// Since this client mimics partial - networking only - functionality of a full node, some of the telemetry-related
		// code paths are executed. To avoid those messages interfering with the telemetry data collected, a non-nil telemetry
		// module that NOOPs (per the configs above) is injected.
		telemetryM, err := telemetry.Create(runtimeMgr)
		if err != nil {
			log.Fatalf("[ERROR] Failed to create NOOP telemetry module: " + err.Error())
		}
		telemetryMod := telemetryM.(modules.TelemetryModule)

		_ = shared.CreateBusWithOptionalModules(runtimeMgr, nil, p2pMod, nil, consensusMod, telemetryMod) // TODO: refactor using the `WithXXXModule()` pattern accepting a slice of IntegratableModule

		p2pMod.Start()
	})
}