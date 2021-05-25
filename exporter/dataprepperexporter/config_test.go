package dataprepperexporter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"path"
	"testing"
	"time"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Exporters[typeStr] = factory
	cfg, err := configtest.LoadConfigFile(t, path.Join(".", "testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	t.Run("Validate", func(t *testing.T) {
		assert.NoError(t, cfg.Validate())
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		defaultExporter := cfg.Exporters[config.NewID(typeStr)]
		assert.Equal(t, defaultExporter, factory.CreateDefaultConfig())
	})

	t.Run("OpenSearch", func(t *testing.T) {
		opensearchExporter := cfg.Exporters[config.NewIDWithName(typeStr, "opensearch")]
		assert.Equal(t, opensearchExporter,
			&Config{
				ExporterSettings: config.NewExporterSettings(config.NewIDWithName(typeStr, "opensearch")),
				RetrySettings: exporterhelper.RetrySettings{
					Enabled:         true,
					InitialInterval: 10 * time.Second,
					MaxInterval:     1 * time.Minute,
					MaxElapsedTime:  10 * time.Minute,
				},
				QueueSettings: exporterhelper.QueueSettings{
					Enabled:      true,
					NumConsumers: 2,
					QueueSize:    10,
				},
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Headers: map[string]string{
						"can you have a . here?": "F0000000-0000-0000-0000-000000000000",
						"header1":                "234",
						"another":                "somevalue",
					},
					Endpoint: "https://1.2.3.4:1234",
					TLSSetting: configtls.TLSClientSetting{
						TLSSetting: configtls.TLSSetting{
							CAFile:   "/var/lib/mycert.pem",
							CertFile: "certfile",
							KeyFile:  "keyfile",
						},
						Insecure: true,
					},
					ReadBufferSize:  123,
					WriteBufferSize: 345,
					Timeout:         time.Second * 10,
				},
			})
	})

	t.Run("AWS", func(t *testing.T) {
		awsExporter := cfg.Exporters[config.NewIDWithName(typeStr, "aws")]
		expAWSExporterConfig := factory.CreateDefaultConfig().(*Config)
		expAWSExporterConfig.ExporterSettings = config.NewExporterSettings(config.NewIDWithName(typeStr, "aws"))
		expAWSExporterConfig.Endpoint = "accountId.dataprepper.us-east-1.es.aws.com"
		expAWSExporterConfig.AWSAuthConfig = AWSAuthConfig{
			PipelineArn: "arn:aws:es::123456789012:es/dataprepper/pipeline-name",
			SigV4Config: SigV4Config{
				Region: "us-east-1",
				RoleArn: "arn:aws:iam::123456789012:role/test-role",
			},
		}
		assert.Equal(t, awsExporter, expAWSExporterConfig)
	})
}

func TestConfigValidateError(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name        string
		pipelineArn string
		region      string
		roleArn     string
	}{
		{
			name: "Invalid pipelineArn",
			pipelineArn: "arn:",
		},
		{
			name: "Invalid roleArn",
			pipelineArn: "arn:aws:es::123456789012:es/dataprepper/pipeline-name",
			region: "us-east-1",
			roleArn: "arn:",
		},
		{
			name: "Missing region",
			pipelineArn: "arn:aws:es::123456789012:es/dataprepper/pipeline-name",
			roleArn: "arn:aws:iam::123456789012:role/test-role",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := factory.CreateDefaultConfig().(*Config)
			cfg.AWSAuthConfig = AWSAuthConfig{
				PipelineArn: test.pipelineArn,
				SigV4Config: SigV4Config{
					Region: test.region,
					RoleArn: test.roleArn,
				},
			}

			assert.Error(t, cfg.Validate())
		})
	}
}
