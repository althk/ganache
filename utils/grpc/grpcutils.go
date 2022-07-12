package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	hpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               2 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

var kacp = keepalive.ClientParameters{
	Time:                8 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             2 * time.Second, // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,            // send pings even without active streams
}

// ZPagesAddr declares the address that will serve
// OpenCensus zpages handler.
var ZPagesAddr = "0.0.0.0:3902"

type TLSConfig struct {
	CertFilePath     string
	KeyFilePath      string
	ClientCAFilePath string
	RootCAFilePath   string
	SkipTLS          bool
	NoClientCert     bool
}

type GRPCServerConfig struct {
	*TLSConfig
	SkipReflection   bool
	SkipHealthServer bool
	SkipZPages       bool
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
	cfg := &tls.Config{
		ClientAuth: tls.NoClientCert,
	}
	if c.NoClientCert {
		return cfg, nil
	}
	tlsKeyPair, err := tls.LoadX509KeyPair(c.CertFilePath, c.KeyFilePath)
	if err != nil {
		return nil, err
	}
	cfg.Certificates = []tls.Certificate{tlsKeyPair}
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
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}, nil
}

func GetGRPCDialOpts(tlsCfg *TLSConfig) ([]grpc.DialOption, error) {
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		return nil, err
	}
	creds, err := tlsCfg.Creds()
	if err != nil {
		return nil, err
	}
	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	}, nil

}

func getServerInterceptorChain() grpc.ServerOption {
	logger := zerolog.New(os.Stdout)
	return middleware.WithUnaryServerChain(
		tags.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(logger)),
		otelgrpc.UnaryServerInterceptor(),
	)
}

func NewGRPCServer(grpcCfg *GRPCServerConfig) (*grpc.Server, error) {
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		return nil, err
	}
	serverOpts, err := GetGRPCServerOpts(grpcCfg.TLSConfig)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(serverOpts...)

	// register other servers
	if !grpcCfg.SkipHealthServer {
		h := health.NewServer()
		hpb.RegisterHealthServer(s, h)
		h.Resume()
	}
	if !grpcCfg.SkipReflection {
		reflection.Register(s)
	}

	if !grpcCfg.SkipZPages {
		go func() {
			mux := http.NewServeMux()
			zpages.Handle(mux, "/debug")

			if err := http.ListenAndServe(ZPagesAddr, mux); err != nil {
				log.Fatal().Err(err).Msg("Failed to start metrics handler")
			}
		}()
	}

	return s, nil
}

func OTelTraceProvider(svcName string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(svcName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "ganache_otelcoll_1:4317", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}
