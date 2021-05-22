package dataprepperexporter

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/arn"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

type SigV4Config struct {
	// Region is the AWS region for AWS SigV4.
	Region string `mapstructure:"region"`

	// Amazon Resource Name (ARN) of a role to assume. Optional.
	RoleArn string `mapstructure:"role_arn"`
}

type AWSAuthConfig struct {
	PipelineArn string `mapstructure:"pipeline_arn"`
	SigV4Config SigV4Config `mapstructure:"sigv4"`
}

// Config defines configuration for OTLP/HTTP exporter.
type Config struct {
	config.ExporterSettings       `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	confighttp.HTTPClientSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings  `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings  `mapstructure:"retry_on_failure"`

	// The URL to send traces to. If omitted the Endpoint + "/v1/traces" will be used.
	TracesEndpoint string `mapstructure:"traces_endpoint"`

	// The URL to send metrics to. If omitted the Endpoint + "/v1/metrics" will be used.
	MetricsEndpoint string `mapstructure:"metrics_endpoint"`

	// The URL to send logs to. If omitted the Endpoint + "/v1/logs" will be used.
	LogsEndpoint string `mapstructure:"logs_endpoint"`

	// The compression key for supported compression types within
	// collector. Currently the only supported mode is `gzip`.
	Compression string `mapstructure:"compression"`

	AWSAuthConfig AWSAuthConfig `mapstructure:"aws_auth"`
}

var _ config.Exporter = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg Config) Validate() error {
	if HasAWSAuth(cfg) {
		if err := cfg.AWSAuthConfig.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func HasAWSAuth(cfg Config) bool {
	return cfg.AWSAuthConfig != AWSAuthConfig{}
}

func HasSigV4(awsAuthConfig AWSAuthConfig) bool {
	return awsAuthConfig.SigV4Config != SigV4Config{}
}

func (sigV4Config SigV4Config) Validate() error {
	return nil
}

func (awsAuth AWSAuthConfig) Validate() error {
	var pipelineArn = awsAuth.PipelineArn
	if !arn.IsARN(pipelineArn) {
		return fmt.Errorf("invalid pipeline_arn: %s", pipelineArn)
	}
	if HasSigV4(awsAuth) {
		if err := awsAuth.SigV4Config.Validate(); err != nil {
			return err
		}
	}
	return nil
}
