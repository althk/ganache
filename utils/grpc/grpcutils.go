package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type TLSConfig struct {
	CertFilePath     string
	KeyFilePath      string
	ClientCAFilePath string
	RootCAFilePath   string
	SkipTLS          bool
}

type GRPCServerConfig struct {
	*TLSConfig
	EnableReflection   bool
	EnableHealthServer bool
}

func (c *TLSConfig) Creds() (credentials.TransportCredentials, error) {
	if c.SkipTLS {
		return insecure.NewCredentials(), nil
	}
	// init new tls config and load the cert
	cfg, err := c.newTLS()
	if err != nil {
		return nil, err
	}
	// if client ca is set, load it and enable client
	// verification
	if err = c.setClientCAs(cfg); err != nil {
		return nil, err
	}
	// if root ca is set, load it to enable server
	// verification
	if err = c.setRootCAs(cfg); err != nil {
		return nil, err
	}

	return credentials.NewTLS(cfg), nil
}

func (c *TLSConfig) newTLS() (*tls.Config, error) {
	tlsKeyPair, err := tls.LoadX509KeyPair(c.CertFilePath, c.KeyFilePath)
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{tlsKeyPair},
		ClientAuth:   tls.NoClientCert,
	}
	return cfg, nil
}

func (c *TLSConfig) setRootCAs(cfg *tls.Config) error {
	if c.RootCAFilePath == "" {
		return nil
	}
	certPool, err := newCertPool(c.RootCAFilePath)
	if err != nil {
		return err
	}
	cfg.RootCAs = certPool

	return nil
}

func (c *TLSConfig) setClientCAs(cfg *tls.Config) error {
	if c.ClientCAFilePath == "" {
		return nil
	}
	certPool, err := newCertPool(c.ClientCAFilePath)
	if err != nil {
		return err
	}
	cfg.ClientCAs = certPool
	cfg.ClientAuth = tls.RequireAndVerifyClientCert

	return nil
}

// newCertPool creates a new CertPool and appends the cert
// at the given path to the pool.
func newCertPool(caPath string) (*x509.CertPool, error) {
	pemData, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("error adding CA cert to pool for %s", caPath)
	}
	return certPool, nil
}

func GetGRPCServerOpts(tlsCfg *TLSConfig) ([]grpc.ServerOption, error) {
	creds, err := tlsCfg.Creds()
	if err != nil {
		return nil, err
	}
	return []grpc.ServerOption{
		getServerInterceptorChain(),
		grpc.Creds(creds),
	}, nil
}

func getServerInterceptorChain() grpc.ServerOption {
	logger := zerolog.New(os.Stdout)
	return middleware.WithUnaryServerChain(
		tags.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(logger)),
	)
}

func NewGRPCServer(grpcCfg *GRPCServerConfig) (*grpc.Server, error) {
	serverOpts, err := GetGRPCServerOpts(grpcCfg.TLSConfig)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(serverOpts...)

	// register other servers
	if grpcCfg.EnableHealthServer {
		h := health.NewServer()
		hpb.RegisterHealthServer(s, h)
	}
	if grpcCfg.EnableReflection {
		reflection.Register(s)
	}

	return s, nil
}
