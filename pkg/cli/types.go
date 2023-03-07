package cli

type runOpts struct {
	verbose   bool
	certPath  string
	keyPath   string
	httpAddr  string
	httpsAddr string
}

func (opts *runOpts) customCAConfigured() bool {
	return opts.certPath != "" && opts.keyPath != ""
}

func newRunOpts() *runOpts {
	return &runOpts{}
}
