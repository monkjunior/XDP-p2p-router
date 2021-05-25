package cmd

import (
	"github.com/spf13/cobra"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

var (
	device string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the router",
	Run:   execStartCmd,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&device, "device", "wlp8s0", "network interface that you want to attach this program to it")
}

func execStartCmd(startCmd *cobra.Command, args []string) {
	packetCapture.Start(device)
}
