package cli

import (
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/node"
	flag "github.com/spf13/pflag"

	"github.com/iotaledger/wasp/plugins/banner"
)

// PluginName is the name of the CLI plugin.
const PluginName = "CLI"

var (
	printVersion bool
)

func Init() *node.Plugin {
	flag.BoolVarP(&printVersion, "version", "v", false, "Prints the Wasp version")

	Plugin := node.NewPlugin(PluginName, node.Enabled)
	Plugin.Events.Init.Attach(events.NewClosure(onInit))
	return Plugin
}

func onAddPlugin(name string, status int) {
	AddPluginStatus(node.GetPluginIdentifier(name), status)
}

func onInit(*node.Plugin) {
	for name, plugin := range node.GetPlugins() {
		onAddPlugin(name, plugin.Status)
	}
	node.Events.AddPlugin.Attach(events.NewClosure(onAddPlugin))

	flag.Usage = printUsage

	if printVersion {
		fmt.Println(banner.AppName + " " + banner.AppVersion)
		os.Exit(0)
	}
}
