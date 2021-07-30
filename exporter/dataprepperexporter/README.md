# Data Prepper Exporter

Exports traces and/or metrics via HTTP using [OTLP](
https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md)
format to [Data Prepper](https://github.com/opendistro-for-elasticsearch/data-prepper).

## Configuration

The following settings are required:

- `endpoint` (no default): The target base URL to send data to (e.g.: https://example.com:55681).
  To send each trace signal a corresponding path "/v1/traces" will be added.

The following settings can be optionally configured:

- `traces_endpoint` (no default): The target URL to send trace data to (e.g.: https://example.com:55681/v1/traces).
   If this setting is present the the `endpoint` setting is ignored for traces.

- `insecure` (default = false): when set to true disables verifying the server's
  certificate chain and host name. The connection is still encrypted but server identity
  is not verified.
- `ca_file` path to the CA cert. For a client this verifies the server certificate. Should
  only be used if `insecure` is set to false.
- `cert_file` path to the TLS cert to use for TLS required connections. Should
  only be used if `insecure` is set to false.
- `key_file` path to the TLS key to use for TLS required connections. Should
  only be used if `insecure` is set to false.

- `timeout` (default = 30s): HTTP request time limit. For details see https://golang.org/pkg/net/http/#Client
- `read_buffer_size` (default = 0): ReadBufferSize for HTTP client.
- `write_buffer_size` (default = 512 * 1024): WriteBufferSize for HTTP client.

The full list of settings exposed for this exporter are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).

## Use cases
### OpenSearch

```yaml
exporters:
  otlphttp:
    endpoint: https://example.com:55681/v1/traces
```
