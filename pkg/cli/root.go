package cli

import (
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "should every proxy request be logged to stdout")
	runCmd.Flags().StringVarP(&opts.certPath, "certpath", "", "", "path to base64-encoded CA cert in PEM format")
	runCmd.Flags().StringVarP(&opts.keyPath, "keypath", "", "", "path to base64-encoded CA private key")
	runCmd.Flags().StringVarP(&opts.httpAddr, "httpaddr", "", ":3129", "proxy http listen address")
	runCmd.Flags().StringVarP(&opts.httpsAddr, "httpsaddr", "", ":3128", "proxy https listen address")

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "simpleproxy",
	Short: "a simple HTTP proxy built on top of elazarl/goproxy",
}

var (
	opts = newRunOpts()
)
