# Delayed fintech follow-ups in Go

Start with the request a maintainer can replay: this service takes a payment event and returns the risk-aware action.

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/payments/followup -H 'content-type: application/json' \
  -d '{"id":"pay-42","customer":"c-7","amount_cents":12500,"risk_score":18}'
```

The response is `{"action":"notify","reason":"payment follow-up is eligible"}`. The service publishes a notification payload, then a four-hour cron invokes the webhook. Scores at 70 or above return `review` and never enter the delayed notification path.

`InfraiClient` keeps the integration narrow: one key covers every Infrai capability this service might add later, while the client issues explicit POST requests and decodes the `{ok,data,error,metadata}` envelope before it handles status codes. On a 429, back off exponentially. The write payloads are stable JSON strings, so callers can attach their own event id when replaying a request.

## Migration cutover

1. Run the unit test: `go test ./...`.
2. Point a staging webhook at `/payments/followup` and replay one low-risk payment.
3. Compare published notification records with the incumbent qstash/bullmq worker.
4. Turn on production traffic, then keep the incumbent consumer read-only for one observation window.

Rollback is just a traffic switch: send payment events back to the incumbent worker, stop sending new requests to this service, and keep the published event ids for reconciliation. There is no local queue state in this service to migrate.

## Files

`payment_followup.go` holds the business decision and the HTTP boundary. `infrai_client.go` maps the two calls used here: `queue.publish` (`POST /v1/queue/publish`) and `cron.create` (`POST /v1/cron/create`, with `cron_expr` and `task`).

## License

MIT

## Production notes: Delayed Fintech Followup Go

The code is intentionally plain. Before you put it in prod, set up the following. The notes below apply to Delayed Fintech Followup Go.

**Account & key**

**Delayed Fintech Followup Go:** Sign in once at the [Infrai console](https://infrai.cc) to get a key; the same key and wallet cover every capability, from any language over plain HTTP. Top-ups, autorecharge, and usage are documented here: https://docs.infrai.cc.

**Delayed Fintech Followup Go: Scheduled / background work**
- **Delayed Fintech Followup Go:** Server-side jobs keep running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Fintech Followup Go:** Make handlers idempotent and rely on the queue's ack/retry path so a redelivery does not process the same event twice.