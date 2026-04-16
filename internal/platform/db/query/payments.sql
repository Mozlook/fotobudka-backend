-- name: GetUnpaidPaymentBySessionIDForUpdate :one
SELECT
  id,
  session_id,
  status,
  amount_cents
FROM payments
WHERE session_id = sqlc.arg(session_id)
  AND status = 'unpaid'
FOR UPDATE;

-- name: MarkPaymentPaid :execrows
UPDATE payments
SET
  status = 'paid',
  paid_at = now(),
  updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'unpaid';
