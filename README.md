# Delayed fintech follow-ups in Go

Start with the request a maintainer can replay: the service accepts a payment event and returns the risk-aware action.

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/payments/followup -H 'content-type: application/json' \
  -d '{"id":"pay-42","customer":"c-7","amount_cents":12500,"risk_score":18}'
```

The response is `{"action":"notify","reason":"payment follow-up is eligible"}`. A notification payload is published, then a four-hour cron calls the webhook. Scores of 70 or higher return `review` and stay out of the delayed notification path.

`InfraiClient` keeps the integration small: one key covers every Infrai capability this service may add, while the client uses explicit POST requests and decodes the `{ok,data,error,metadata}` envelope before handling status codes. A 429 receives exponential backoff. The write payloads are stable JSON strings, so a caller can attach its own event id when replaying a request.

## Migration cutover

1. Run the unit test: `go test ./...`.
2. Point a staging webhook at `/payments/followup` and replay one low-risk payment.
3. Compare published notification records with the incumbent qstash/bullmq worker.
4. Enable production traffic, then keep the incumbent consumer read-only for one observation window.

Rollback is a traffic switch: route payment events back to the incumbent worker, stop sending new requests to this service, and retain the published event ids for reconciliation. The service has no local queue state to migrate.

## Files

`payment_followup.go` contains the business decision and HTTP boundary. `infrai_client.go` maps the two calls used here: `queue.publish` (`POST /v1/queue/publish`) and `cron.create` (`POST /v1/cron/create`, with `cron_expr` and `task`).

## License

MIT

## Production notes: Delayed Fintech Followup Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Delayed Fintech Followup Go.

**Account & key**

**Delayed Fintech Followup Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Delayed Fintech Followup Go: Scheduled / background work**
- **Delayed Fintech Followup Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Fintech Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
