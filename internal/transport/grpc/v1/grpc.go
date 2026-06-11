package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"strings"
	"time"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/logger"
	mapper_err_grpc "github.com/Weit145/gateway-startup/internal/transport/grpc/v1/err"

	GRPCauth "github.com/weit145/proto-startup/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type TLSSettings struct {
	CertPEM    []byte
	KeyPEM     []byte
	ServerCA   []byte
	ServerName string
}

type Client struct {
	api  GRPCauth.AuthClient
	conn *grpc.ClientConn
}

func New(addr string, tlsSettings TLSSettings) (*Client, error) {
	if addr == "" {
		slog.Error("Empty address provided for gRPC client")
		return nil, fmt.Errorf("%w", errs.ErrInternal)
	}

	transportCredentials, err := NewTLSCredentials(tlsSettings)
	if err != nil {
		return nil, fmt.Errorf("create auth gRPC TLS credentials: %w", err)
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		slog.Error("Failed to create gRPC client connection", "error", err)
		return nil, fmt.Errorf("create auth gRPC client: %w", err)
	}

	conn.Connect()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}

		if !conn.WaitForStateChange(ctx, state) {
			slog.Error("gRPC connection did not become ready", "state", state)
			return nil, fmt.Errorf("%w", errs.ErrInternal)
		}
	}

	grpcClient := GRPCauth.NewAuthClient(conn)

	return &Client{
		api:  grpcClient,
		conn: conn,
	}, nil
}

func NewTLSCredentials(settings TLSSettings) (credentials.TransportCredentials, error) {
	certificate, err := tls.X509KeyPair(settings.CertPEM, settings.KeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse gRPC client certificate: %w", err)
	}

	serverCAs := x509.NewCertPool()
	if ok := serverCAs.AppendCertsFromPEM(settings.ServerCA); !ok {
		return nil, fmt.Errorf("parse gRPC server CA certificate")
	}

	serverName := strings.TrimSpace(settings.ServerName)
	if serverName == "" {
		return nil, fmt.Errorf("gRPC TLS server name is required")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      serverCAs,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}), nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CheckHealth(ctx context.Context) (bool, error) {
	resp, err := c.api.CheckHealth(ctx, &GRPCauth.Empty{})
	if err != nil {
		err = mapper_err_grpc.MapErr(err)
		slog.DebugContext(ctx, "Failed to check health via gRPC", "error", err)
		return false, logger.WrapError(ctx, err)
	}

	return resp.GetHealth(), nil
}
