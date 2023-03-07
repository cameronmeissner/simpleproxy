package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "simpleproxy",
	Short: "a simple HTTP proxy built on top of elazarl/goproxy",
	Long: "a simple HTTP proxy built on top of elazarl/goproxy"
}