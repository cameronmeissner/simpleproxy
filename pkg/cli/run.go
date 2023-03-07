package cli

import (
	"github.com/spf13/cobra"
	"log"
	"flag"
)

type runOpts struct {
	verbose string
	cacert []byte
	caKey []byte
	httpAddr string
	httpsAddr string
}

func (opts *runOpts) customCAConfigured() bool {
	return opts.cacert != "" && opts.cakey != ""
}

var runCmd = &cobra.Command{
	Use: "run",
	Short: "run simpleproxy on specified ports",
	Long: "run simpleproxy on specified HTTP and HTTPS ports, optionally with a custom CA certificate & private key",
	Run: runProxy,
}

func init() {
	rootCmd.Flags()
}

func runProxy(cmd *cobra.Command, args []string) {
	opts := createRunOpts()
	if opts.verbose {
		log.Printf("simpleproxy will listen on http address: %q, https address: %q", opts.httpAddr, opts.httpsAddr)
	}
	if opts.customCAConfigured() {
		if opts.verbose {
			log.Printf("using custom CA certificate in %q and custom CA private key in %q", opts.cacert, opts.cakey)
		}
		if err := loadCACertAndPrivateKey(opts.cacert, opts.cakey); err != nil {
			panic(err)
		}
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose
	errChannel := make(chan error)

	go func() {
		err := http.ListenAndServe(opts.httpAddr, proxy)
		errChannel <- err
	}()
	go func() {
		err := http.ListenAndServe(opts.httpsAddr, proxy)
		errChannel <- err
	}()

	err := <-errChannel
	log.Fatal(err)
}

func createRunOpts() *runOpts {
	verbose := runCmd.Flags().BoolVarP("v", false, "should every proxy request be logged to stdout")
	cacert := runCmd.Flags().StringVarP("cacert", "", "path to base64-encoded CA cert in PEM format")
	cakey := flag.String("cakey", "", "path to base64-encoded CA private key in PEM format")
	httpAddr := flag.String("httpaddr", ":3129", "proxy http listen address")
	httpsAddr := flag.String("httpsaddr", ":3128", "proxy https listen address")
	flag.Parse()

	return &runOpts{
		verbose: *verbose,
		cacert: *cacert,
		cakey: *cakey,
		httpAddr: *httpAddr,
		httpsAddr: *httpsAddr,
	}
}

func loadCACertAndPrivateKey(certPath, keyPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	ca, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = ca
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&ca)}
	return nil
}
