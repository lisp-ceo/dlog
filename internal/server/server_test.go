package server

import (
	"context"
	"io/ioutil"
	"net"
	"testing"
	"time"

	api "github.com/lisp-ceo/dlog/api/v1"
	"github.com/lisp-ceo/dlog/internal/config"
	"github.com/lisp-ceo/dlog/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, root api.LogClient, unauthorized api.LogClient, cfg *Config){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundary fails":                    testConsumePastBoundary,
		"unauthorized connection fails":                      testUnauthorized,
	} {
		t.Run(scenario, func(t *testing.T) {
			root, unauthorized, cfg, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, root, unauthorized, cfg)
		})
	}
}

func testUnauthorized(t *testing.T, _, unauthorized api.LogClient, cfg *Config) {
	ctx := context.Background()
	produce, err := unauthorized.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("I want in."),
		},
	})
	require.Nil(t, produce)

	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	require.Equal(t, wantCode, gotCode)

	consume, err := unauthorized.Consume(ctx, &api.ConsumeRequest{
		Offset: 0,
	})
	require.Nil(t, consume)

	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	require.Equal(t, gotCode, wantCode)
}

func testProduceConsume(t *testing.T, client, _ api.LogClient, _ *Config) {
	ctx := context.Background()
	ctx, _ = context.WithDeadline(ctx, time.Now().Add(10 * time.Second))

	want := &api.Record{
		Value: []byte("hello world"),
	}

	produce, err := client.Produce(ctx, &api.ProduceRequest{
		Record: want,
	})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset,
	})
	require.NoError(t, err)
	require.Equal(t, want, consume.Record)
}

func testProduceConsumeStream(t *testing.T, client, _ api.LogClient, _ *Config) {
	ctx := context.Background()

	records := []*api.Record{
		{
			Value:  []byte("first message"),
			Offset: 0,
		},
		{
			Value:  []byte("second message"),
			Offset: 1,
		},
	}
	{
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)

		for offset, record := range records {
			err = stream.Send(&api.ProduceRequest{
				Record: record,
			})
			require.NoError(t, err)

			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, uint64(offset), res.Offset)
		}
	}
	{
		stream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{
			Offset: 0,
		})
		require.NoError(t, err)

		for _, record := range records {
			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, res.Record, record)
		}
	}
}

func testConsumePastBoundary(t *testing.T, client, _ api.LogClient, _ *Config) {
	ctx := context.Background()

	produce, err := client.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("hello world"),
		},
	})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset + 1,
	})
	require.Nil(t, consume)

	got := status.Code(err)
	want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())

	assert.Equal(t, want, got)
}

func setupTest(t *testing.T, fn func(*Config)) (api.LogClient, api.LogClient, *Config, func()) {
	t.Helper()

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	// client tls setup
	newClient := func(crtPath, keyPath string) (
		*grpc.ClientConn,
		api.LogClient,
		[]grpc.DialOption,
	) {
		tlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
			CertFile: crtPath,
			KeyFile:  keyPath,
			CAFile:   config.CAFile,
			Server:   false,
		})
		require.NoError(t, err)

		tlsCreds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(tlsCreds),
		}
		conn, err := grpc.Dial(l.Addr().String(), opts...)
		require.NoError(t, err)

		client := api.NewLogClient(conn)
		return conn, client, opts
	}

	var rootConn *grpc.ClientConn
	var rootClient api.LogClient
	rootConn, rootClient, _ = newClient(
		config.RootClientCertFile,
		config.RootClientKeyFile,
	)

	var unauthorizedConn *grpc.ClientConn
	var unauthorizedClient api.LogClient
	unauthorizedConn, unauthorizedClient, _ = newClient(
		config.UnauthorizedCertFile,
		config.UnauthorizedKeyFile,
	)

	// server tls setup
	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(serverTLSConfig)
	dir, err := ioutil.TempDir("", "server-test")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	cfg := &Config{
		l: clog,
	}

	if fn != nil {
		fn(cfg)
	}

	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	go func() {
		_ = server.Serve(l)
	}()

	return rootClient, unauthorizedClient, cfg, func() {
		server.Stop()

		_ = rootConn.Close()
		_ = unauthorizedConn.Close()
		_ = l.Close()

	}

}
