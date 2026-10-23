# Mock Transactions API

Base URL: `http://localhost:18080`. All timestamps are UTC RFC 3339 strings. All amounts use integer cents and currency is USD.

## Authentication

Every `/v1/` request requires this header:

```http
Authorization: Bearer local-test-token
```

Missing or invalid credentials return `401 Unauthorized` with `{"error":"unauthorized"}`.

## List account transactions

`GET /v1/accounts/{account_id}/transactions?limit=4&cursor=...`

`limit` is optional, defaults to `4`, and must be an integer from `1` through `20`. Omit `cursor` to start at the beginning. A returned cursor is an opaque continuation token; pass it unchanged to fetch the next page. Results retain the fixture's deterministic order, which is not guaranteed to be chronological. A page item contains only `id`, `status`, and `created_at`.

```json
{
  "items": [
    {"id": "tx_001", "status": "COMPLETED", "created_at": "2026-10-01T00:00:00Z"}
  ],
  "next_cursor": "cursor_4"
}
```

`next_cursor` is `null` on the last page. Known accounts are `acct_001`, `acct_002`, and `acct_empty`; the latter has no transactions. Unknown accounts return `404 Not Found`. Invalid limits and cursors return `400 Bad Request`.

## Get transaction details

`GET /v1/transactions/{transaction_id}`

Returns the transaction record:

```json
{
  "id": "tx_001",
  "account_id": "acct_001",
  "status": "COMPLETED",
  "amount_cents": 1599,
  "currency": "USD",
  "created_at": "2026-10-01T00:00:00Z"
}
```

Unknown transaction IDs return `404 Not Found`.

## Errors

Errors use a JSON object with an `error` string. Status codes are `400` for malformed pagination parameters, `401` for failed authentication, `404` for unknown accounts or transaction IDs, `429` for rate limiting, and `503` for temporary service unavailability.

## Local diagnostics

- `GET /healthz` returns `{"status":"ok"}`.
- `POST /_admin/reset` resets transient failure counters and clears recorded requests.
- `GET /_admin/requests` returns recent request method, path, and status code records (up to 100).

Admin routes are unauthenticated and intended only for localhost practice. The server listener binds to `127.0.0.1`.

## Scenarios

Set `MOCK_SCENARIO=normal` (default) for successful valid requests and normal validation errors. Set `MOCK_SCENARIO=transient` to make the first detail request for `tx_003` return `503`, and the first detail request for `tx_004` return `429` with `Retry-After: 1`. Each selected failure occurs once after startup or reset, then that detail request succeeds.
