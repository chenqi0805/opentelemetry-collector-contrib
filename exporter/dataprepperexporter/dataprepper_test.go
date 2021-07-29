package dataprepperexporter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dataprepperexporter/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/testutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestInvalidConfig(t *testing.T) {
	config := &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "",
		},
	}
	f := NewFactory()
	params := component.ExporterCreateParams{Logger: zap.NewNop()}
	_, err := f.CreateTracesExporter(context.Background(), params, config)
	require.Error(t, err)
	_, err = f.CreateMetricsExporter(context.Background(), params, config)
	require.Error(t, err)
	_, err = f.CreateLogsExporter(context.Background(), params, config)
	require.Error(t, err)
}

func TestTraceNoBackend(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	exp := startTracesExporter(t, "", fmt.Sprintf("http://%s/v1/traces", addr), "", "")
	td := testdata.GenerateTracesOneSpan()
	assert.Error(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceInvalidUrl(t *testing.T) {
	exp := startTracesExporter(t, "http:/\\//this_is_an/*/invalid_url", "", "", "")
	td := testdata.GenerateTracesOneSpan()
	assert.Error(t, exp.ConsumeTraces(context.Background(), td))

	exp = startTracesExporter(t, "", "http:/\\//this_is_an/*/invalid_url", "", "")
	td = testdata.GenerateTracesOneSpan()
	assert.Error(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceRoundTripOnlyBaseUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/protobuf", r.Header.Get("Content-Type"))
	}
	server := startTestHttpServer(t, "/v1/traces", handler)
	exp := startTracesExporter(t, server.URL, "", "", "")
	td := testdata.GenerateTracesOneSpan()
	assert.NoError(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceRoundTripWithTraceEndpoint(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/protobuf", r.Header.Get("Content-Type"))
	}
	server := startTestHttpServer(t, "/custom/traces", handler)
	exp := startTracesExporter(t, server.URL, server.URL + "/custom/traces", "", "")
	td := testdata.GenerateTracesOneSpan()
	assert.NoError(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceRoundTripAWSEndpointNoSigV4(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-name", r.Header.Get(headerDataPrepper))
		assert.Equal(t, "123456789012.ingest.us-east-1.amazonaws.com", r.Host)
	}
	server := startTestHttpServer(t, "/v1/traces", handler)
	exp := startTracesExporter(t, server.URL, "", "arn:aws:es:us-east-1:123456789012:es/dataprepper/test-name", "")
	td := testdata.GenerateTracesOneSpan()
	assert.NoError(t, exp.ConsumeTraces(context.Background(), td))
}

func TestTraceRoundTripAWSEndpointSigV4(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "TEST_AWS_ACCESS_KEY")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "TEST_AWS_SECRET_ACCESS_KEY")
	t.Cleanup(os.Clearenv)
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header, "X-Amz-Content-Sha256")
		assert.Equal(t, "application/protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-name", r.Header.Get(headerDataPrepper))
		assert.Equal(t, "123456789012.ingest.us-east-1.amazonaws.com", r.Host)
		authStr := r.Header.Get("Authorization")
		assert.Contains(t, authStr, "Credential=TEST_AWS_ACCESS_KEY")
		assert.Contains(t, authStr, "SignedHeaders=host;x-amz-content-sha256;x-amz-date")
		_, err := v4.GetSignedRequestSignature(r)
		assert.NoError(t, err)
	}
	server := startTestHttpServer(t, "/v1/traces", handler)
	exp := startTracesExporter(t, server.URL, "", "arn:aws:es:us-east-1:123456789012:es/dataprepper/test-name", "us-east-1")
	td := testdata.GenerateTracesOneSpan()
	assert.NoError(t, exp.ConsumeTraces(context.Background(), td))
}

func startTracesExporter(t *testing.T, baseURL string, overrideURL string, pipelineArn string, region string) component.TracesExporter {
	factory := NewFactory()
	cfg := createExporterConfig(baseURL, factory.CreateDefaultConfig())
	cfg.TracesEndpoint = overrideURL
	cfg.AWSAuthConfig = AWSAuthConfig{}
	cfg.AWSAuthConfig.PipelineArn = pipelineArn
	cfg.AWSAuthConfig.SigV4Config = SigV4Config{Region: region}
	exp, err := factory.CreateTracesExporter(context.Background(), component.ExporterCreateParams{Logger: zap.NewNop()}, cfg)
	require.NoError(t, err)
	startAndCleanup(t, exp)
	return exp
}

func createExporterConfig(baseURL string, defaultCfg config.Exporter) *Config {
	cfg := defaultCfg.(*Config)
	cfg.Endpoint = baseURL
	cfg.QueueSettings.Enabled = false
	cfg.RetrySettings.Enabled = false
	return cfg
}

func startTestHttpServer(t *testing.T, path string, handlerFunc func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc(path, handlerFunc)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func startAndCleanup(t *testing.T, cmp component.Component) {
	require.NoError(t, cmp.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() {
		require.NoError(t, cmp.Shutdown(context.Background()))
	})
}

func TestErrorResponses(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	errMsgPrefix := fmt.Sprintf("error exporting items, request to http://%s/v1/traces responded with HTTP Status Code ", addr)

	tests := []struct {
		name           string
		responseStatus int
		responseBody   *status.Status
		err            error
		isPermErr      bool
		headers        map[string]string
	}{
		{
			name:           "400",
			responseStatus: http.StatusBadRequest,
			responseBody:   status.New(codes.InvalidArgument, "Bad field"),
			isPermErr:      true,
		},
		{
			name:           "404",
			responseStatus: http.StatusNotFound,
			err:            fmt.Errorf(errMsgPrefix + "404"),
		},
		{
			name:           "419",
			responseStatus: http.StatusTooManyRequests,
			responseBody:   status.New(codes.InvalidArgument, "Quota exceeded"),
			err: exporterhelper.NewThrottleRetry(
				fmt.Errorf(errMsgPrefix+"429, Message=Quota exceeded, Details=[]"),
				time.Duration(0)*time.Second),
		},
		{
			name:           "503",
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   status.New(codes.InvalidArgument, "Server overloaded"),
			err: exporterhelper.NewThrottleRetry(
				fmt.Errorf(errMsgPrefix+"503, Message=Server overloaded, Details=[]"),
				time.Duration(0)*time.Second),
		},
		{
			name:           "503-Retry-After",
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   status.New(codes.InvalidArgument, "Server overloaded"),
			headers:        map[string]string{"Retry-After": "30"},
			err: exporterhelper.NewThrottleRetry(
				fmt.Errorf(errMsgPrefix+"503, Message=Server overloaded, Details=[]"),
				time.Duration(30)*time.Second),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serveMux := http.NewServeMux()
			serveMux.HandleFunc("/v1/traces", func(writer http.ResponseWriter, request *http.Request) {
				for k, v := range test.headers {
					writer.Header().Add(k, v)
				}
				writer.WriteHeader(test.responseStatus)
				if test.responseBody != nil {
					msg, err := proto.Marshal(test.responseBody.Proto())
					require.NoError(t, err)
					_, err = writer.Write(msg)
					require.NoError(t, err)
				}
			})
			srv := http.Server{
				Addr:    addr,
				Handler: serveMux,
			}
			ln, err := net.Listen("tcp", addr)
			require.NoError(t, err)
			go func() {
				_ = srv.Serve(ln)
			}()

			cfg := &Config{
				ExporterSettings: config.NewExporterSettings(config.NewID(typeStr)),
				TracesEndpoint:   fmt.Sprintf("http://%s/v1/traces", addr),
				// Create without QueueSettings and RetrySettings so that ConsumeTraces
				// returns the errors that we want to check immediately.
			}
			exp, err := createTracesExporter(context.Background(), component.ExporterCreateParams{Logger: zap.NewNop()}, cfg)
			require.NoError(t, err)

			traces := pdata.NewTraces()
			err = exp.ConsumeTraces(context.Background(), traces)
			assert.Error(t, err)

			if test.isPermErr {
				assert.True(t, consumererror.IsPermanent(err))
			} else {
				assert.EqualValues(t, test.err, err)
			}

			srv.Close()
		})
	}
}
