package remote

import (
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/drmaa2os"
	"github.com/dgruber/drmaa2os/pkg/jobtracker/remote/client"
	genclient "github.com/dgruber/drmaa2os/pkg/jobtracker/remote/client/generated"
	"github.com/dgruber/wfl"
)

type BasicAuthConfig struct {
	User     string
	Password string
}

type Config struct {
	// Server is the URL of the DRMAA2 server including port, like
	// http://localhost:8088
	Server string
	// Path is the path to the DRMAA2 server, like /jobserver/jobmanagement
	Path string
	// BasicAuth uses username and password for authentication if set
	BasicAuth *BasicAuthConfig
	// JobSessionName is the name of the DRMAA2 job session
	JobSessionName string
	// JobSessionDBFile is the path to the job session database file
	JobSessionDBFile string
	// DefaultTemplate is the job template which is used for all jobs,
	// when not explicitly set in the job submission.
	DefaultTemplate drmaa2interface.JobTemplate
}

// NewRemoteContextByCfg creates a wfl Context which executes tasks
// remotely on a DRMAA2 server.
func NewRemoteContextByCfg(cfg Config) *wfl.Context {

	if cfg.JobSessionDBFile == "" {
		cfg.JobSessionDBFile = wfl.TmpFile()
	}

	if cfg.Server == "" {
		return &wfl.Context{
			CtxCreationErr: fmt.Errorf("Server URL is not set"),
			SMType:         wfl.RemoteSessionManager,
			JobSessionName: cfg.JobSessionName,
		}
	}

	var basicAuthProvider *securityprovider.SecurityProviderBasicAuth

	if cfg.BasicAuth != nil {
		var err error
		basicAuthProvider, err = securityprovider.NewSecurityProviderBasicAuth(
			cfg.BasicAuth.User, cfg.BasicAuth.Password)
		if err != nil {
			return &wfl.Context{
				CtxCreationErr: err,
				SMType:         wfl.RemoteSessionManager,
				JobSessionName: cfg.JobSessionName,
			}
		}
	}

	clientTrackerArgs := client.ClientTrackerParams{
		Server: cfg.Server,
		Path:   cfg.Path,
	}

	if basicAuthProvider != nil {
		clientTrackerArgs.Opts = append(clientTrackerArgs.Opts,
			genclient.WithRequestEditorFn(basicAuthProvider.Intercept))
	}

	sm, err := drmaa2os.NewRemoteSessionManager(clientTrackerArgs,
		cfg.JobSessionDBFile)

	return &wfl.Context{
		SM:              sm,
		SMType:          wfl.RemoteSessionManager,
		CtxCreationErr:  err,
		DefaultTemplate: cfg.DefaultTemplate,
		JobSessionName:  cfg.JobSessionName,
	}
}
