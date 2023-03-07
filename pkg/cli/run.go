package cli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run simpleproxy on specified ports",
	Long:  "run simpleproxy on specified HTTP and HTTPS ports, optionally with a custom CA certificate & private key",
	Run:   runProxy,
}

func runProxy(cmd *cobra.Command, args []string) {
	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	// setup signal handling to cancel the context
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM)
		<-signals
		log.Println("received SIGTERM. Terminating...")
		shutdown()
	}()

	log.Printf("simpleproxy will listen on http address: %q, https address: %q", opts.httpAddr, opts.httpsAddr)
	if opts.customCAConfigured() {
		log.Printf("using custom CA certificate in %q and custom CA private key in %q", opts.certPath, opts.keyPath)
		if err := loadCACertAndPrivateKey(opts.certPath, opts.keyPath); err != nil {
			log.Fatal(ctx, err.Error())
		}
	}

	errs := make(chan error)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = opts.verbose

	// run an instance of the proxy on both specified ports
	go func() {
		errs <- http.ListenAndServe(opts.httpAddr, proxy)
	}()
	go func() {
		errs <- http.ListenAndServe(opts.httpsAddr, proxy)
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-errs:
		log.Fatal(ctx, err.Error())
	}
}

func loadCACertAndPrivateKey(certPath, keyPath string) error {
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	certPEM, err := base64.StdEncoding.DecodeString(string(cert))
	if err != nil {
		return err
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	keyPEM, err := base64.StdEncoding.DecodeString(string(key))
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

	injectCustomCA(&ca)
	return nil
}

func injectCustomCA(ca *tls.Certificate) {
	goproxy.GoproxyCa = *ca
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(ca)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(ca)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(ca)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(ca)}
}
