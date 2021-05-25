package dataprepperexporter

import (
	"errors"
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
	// TODO: list all valid regions
	if sigV4Config.Region == "" {
		return errors.New("region cannot be empty")
	}
	var roleArn = sigV4Config.RoleArn
	if roleArn != "" && !arn.IsARN(roleArn) {
		return fmt.Errorf("invalid role_arn: %s", roleArn)
	}
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
