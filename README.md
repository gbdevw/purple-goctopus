# KRAGOC: Kraken REST API Go Client 

A Go client which eases programmatic use of Kraken REST API.

## Principles

- Based only on standard Go libraries
- Pluggable http.Client
- Pluggable nonce generator
- Fully configurable options for the API client
- Typed errors for HTTP errors (HTTPError) and API errors (KrakenAPIClientErrorBundle, ...)
- No over-engineering : The client is here to format requests and process responses not to do validation or logic.
- 2FA supported

## Other documents

- [Dev documentation](DEV.md)
- [Test documentation](TEST.md)

## Basic usage

```go
// Create the client using one of the provided factories
client := NewPublicWithDefaultOptions()

// Call API endpoint
resp, err := client.GetSystemStatus()
...
```

## Public endpoints - Market Data

| ENDPOINT | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | --- | --- |
| Get Server Time | V| V | V |
| Get Server Status | V | V | V |
| Get Asset Info | V | V | V |
| Get Tradable Asset Pairs | V | V | V |
| Get Ticker Information | V | V | V |
| Get OHLC Data | V | V | V |
| Get Order Book | V | V | V |
| Get Recent Trades | V | V | V |
| Get Recent Spread | V | V | V |

## Private endpoints - User Data

| ENDPOINT | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | --- | --- |
| Get Account Balance | V| V | V |
| Get Trade Balance | V | V | V |
| Get Open Orders | V | V | V |
| Get Closed Orders | V | V | V |
| Query Orders Info | V | V | V |
| Get Trades History | V | V | V |
| Query Trades Info| V | V | V |
| Get Open Positions* | V | V | V |
| Get Ledgers Info | V | V | V |
| Query Ledgers | V | V | V |
| Get Trade Volume | V | V | V |
| Request Export Report | V | V | V |
| Get Export Report Status | V | V | V |
| Retrieve Data Export | V | V | V |
| Delete Export Report | V | V | V |

*No market consolidation

## Private endpoints - User Trading

| ENDPOINT | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | --- | --- |
| Add Order | V | V | V |
| Add Order Batch | V | V | V |
| Edit Order | V | V | V |
| Cancel Order | V | V | V |
| Cancel All Orders | V | V | V |
| Cancel All Orders After | V | V | V |
| Cancel Order Batch | V | V | V |

## Private endpoints - User Funding

| ENDPOINT | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | --- | --- |
| Get Deposit Methods  | V | V | V |
| Get Deposit Addresses | V | V | V |
| Get Status of Recent Deposits | V | V | V |
| Get Withdrawal Information | V | V | V |
| Withdraw Funds | V | V | V* |
| Get Status of Recent Withdrawals | V | V | V* |
| Request Withdrawal Cancelation | V | V | V |
| Request Wallet Transfer | V | V | V* |

* These test scenarios can be improved. They are designed to trigger an error from the server.

## Private endpoints - User Staking

| ENDPOINT | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | --- | --- |
| Stake Asset | V | V | V |
| Unstake Asset | V | V | V |
| List of Stakeable Assets | V | V | V |
| Get Pending Staking Transactions | V | V | V |
| List of Staking Transactions | V | V | V |

## Supported security schemes

| SCHEME | IMPLEMENTED | UNIT TESTS | INTEGRATION TESTS |
| --- | --- | ---| --- |
| API Key | V | V | V |
| API Key + OTP | V | V | V |