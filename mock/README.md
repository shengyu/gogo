# Mock Transactions API

A standalone, standard-library-only Go server for local backend assessment practice. This directory is its own Go module and does not import or start the parent application's main service. Fixtures are deterministic and stored in memory; no database or Docker is needed.

Requires Go 1.22 or newer. From the repository root:

```sh
make mock
```

The mock listens on `127.0.0.1:18080`. The parent service remains separate at its default port `8081`. To change the mock's port or enable deterministic transient errors:

```sh
PORT=9090 make mock
make mock-transient
```

Run mock tests from the repository root:

```sh
make mock-test
```

Tests for the parent service run separately from the repository root.

You can also run this module independently with `cd mock && go run .` from the repository root.

Example requests from another terminal or local service:

```sh
curl http://localhost:18080/healthz
curl -H 'Authorization: Bearer local-test-token' \
  'http://localhost:18080/v1/accounts/acct_001/transactions?limit=4'
curl -H 'Authorization: Bearer local-test-token' \
  http://localhost:18080/v1/transactions/tx_001
curl -X POST http://localhost:18080/_admin/reset
curl http://localhost:18080/_admin/requests
```

See [API.md](API.md) for the public contract, authentication, pagination, and scenario behavior. `fixtures.go` contains the server dataset. Expected aggregation fixtures are kept only in the separate server test fixture file.
