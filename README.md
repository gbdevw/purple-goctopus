# Purple Goctopus

Unofficial Golang SDKs and integration utilities for Kraken products.

## Status: Work in progress

Currently, only the REST and the Websocket SDKs for the Kraken spot exchange are fully implemented and have unit tests and/or integration tests. Integration tests are designed to use validation features: No order should be created and no fee should be charged (please, verify). It is recommended to run the tests with a separate account as target as cancel requests, especially the Cancel All Orders ones, can cancel orders.

## Run unit tests only

```
go test -cover -short ./...
```

## Run integration tests

```
export KRAKEN_API_KEY="SECRET"
export KRAKEN_API_SECRET="SECRET"
export KRAKEN_API_OTP="SECRET"
go test -p 1 -cover ./...
```

Hint: The second factor is optional. Use KRAKEN_API_OTP only if you defined a password second factor for your API key.

## Principles

- Based only on standard Go libraries and on some self-developped frameworks (gosette & gowse)
- Fully configurable, all spare parts are visible and customizable
- Optional built-in observability with the OpenTelemetry framework
- All security options provided by Kraken (password second factor) are supported
