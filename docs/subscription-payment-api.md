# Subscription and Payment API

Base URL:

```text
http://localhost:8080/api/v1
```

Authentication:

- All endpoints in this document require `Authorization: Bearer <access_token>`, except the Midtrans webhook.
- `superadmin` can access all schools.
- `admin_sekolah` is restricted to its own school for payment history and checkout.

## 1. List Subscription Plans

Endpoint:

```http
GET /schools/plans
```

Purpose:

- Get all active subscription plans that can be shown in the client pricing page.

Success response example:

```json
{
  "success": true,
  "message": "plans retrieved",
  "data": [
    {
      "id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
      "name": "Pro",
      "description": "Paket untuk sekolah berkembang yang butuh kapasitas lebih besar.",
      "features": [
        "Maks 1500 siswa",
        "1 sekolah",
        "Semua fitur Basic",
        "Laporan operasional lebih besar"
      ],
      "max_students": 1500,
      "monthly_price": 1299000,
      "yearly_price": 12990000
    }
  ]
}
```

## 2. List Schools With Subscription

Endpoint:

```http
GET /schools
```

Query params:

- `search` optional
- `status` optional: `active` or `nonactive`
- `page` optional
- `per_page` optional

Purpose:

- `superadmin`: show all schools with their active/trial subscription.
- `admin_sekolah`: returns only its own school.

Success response example:

```json
{
  "success": true,
  "message": "schools retrieved",
  "data": [
    {
      "id": "4c507ee2-dd3e-41d3-830e-a91c09f2b0f2",
      "name": "Sekolah QA Subscription",
      "status": "active",
      "subscription": {
        "id": "0d4d24e7-c3d5-4ed6-b4c3-90379af7fdfd",
        "school_id": "4c507ee2-dd3e-41d3-830e-a91c09f2b0f2",
        "status": "active",
        "cycle": "month",
        "quantity": 1,
        "price": 1299000,
        "ends_at": "2026-07-08T00:00:00+08:00",
        "plan": {
          "id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
          "name": "Pro",
          "description": "Paket untuk sekolah berkembang yang butuh kapasitas lebih besar.",
          "features": ["Maks 1500 siswa"],
          "max_students": 1500,
          "monthly_price": 1299000,
          "yearly_price": 12990000
        },
        "created_at": "2026-06-08T00:00:00+08:00",
        "updated_at": "2026-06-08T00:00:00+08:00"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

## 3. Get Current Subscription of a School

Endpoint:

```http
GET /schools/{id}/subscription
```

Purpose:

- Get current active or trial subscription for a specific school.

Notes:

- `superadmin` can access any school.
- `admin_sekolah` can access only its own school.

## 4. Change Subscription Directly

Endpoint:

```http
PUT /schools/{id}/subscription
```

Purpose:

- Internal override endpoint for `superadmin`.
- Used when backend/admin needs to replace active subscription directly without checkout.

Request body:

```json
{
  "plan_id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
  "cycle": "month"
}
```

Important rules:

- Only `superadmin` can call this endpoint.
- Downgrade is allowed only when current active student count still fits the target plan quota.

## 5. Create Subscription Checkout

Endpoint:

```http
POST /schools/{id}/subscription/checkout
```

Purpose:

- Create a Midtrans Snap transaction for switching to another paid plan.

Request body:

```json
{
  "plan_id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
  "cycle": "month"
}
```

Business rules:

- Trial/free plan cannot be purchased from checkout.
- Same plan as current active plan will be rejected.
- Downgrade is allowed if active student count still fits the target plan.
- If there is still an unexpired `pending` payment for the school, a new checkout is rejected.

Success response example:

```json
{
  "success": true,
  "message": "checkout created",
  "data": {
    "id": "d55ba687-445e-41d2-b61b-02ea7683e253",
    "school_id": "4c507ee2-dd3e-41d3-830e-a91c09f2b0f2",
    "school_name": "Sekolah QA Subscription",
    "plan_id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
    "plan_name": "Pro",
    "created_by_user_id": "e82c418e-eb64-414d-9463-aab0e197b5fc",
    "status": "pending",
    "cycle": "month",
    "amount": 1299000,
    "currency": "IDR",
    "provider": "midtrans",
    "provider_order_id": "EA-d55ba687-445e-41d2-b61b-02ea7683e253",
    "provider_snap_token": "snap-token",
    "provider_redirect_url": "https://app.sandbox.midtrans.com/snap/v4/redirection/...",
    "expires_at": "2026-06-09T00:47:47+08:00",
    "created_at": "2026-06-08T00:47:47+08:00",
    "updated_at": "2026-06-08T00:47:47+08:00"
  }
}
```

Frontend flow after success:

1. Save `payment_id` from `data.id`.
2. Open `provider_redirect_url`.
3. After payment, poll payment status endpoint.
4. If status becomes `paid`, refresh school subscription.

## 6. Get Payment Status Detail

Endpoint:

```http
GET /schools/{id}/subscription/payments/{payment_id}
```

Purpose:

- Get the latest state of a payment transaction.
- For Midtrans pending transactions, backend will try to sync latest transaction status from the gateway.

Common statuses:

- `pending`
- `paid`
- `failed`
- `expired`
- `cancelled`

Response example:

```json
{
  "success": true,
  "message": "payment retrieved",
  "data": {
    "id": "d55ba687-445e-41d2-b61b-02ea7683e253",
    "school_id": "4c507ee2-dd3e-41d3-830e-a91c09f2b0f2",
    "school_name": "Sekolah QA Subscription",
    "plan_id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
    "plan_name": "Pro",
    "created_by_user_id": "e82c418e-eb64-414d-9463-aab0e197b5fc",
    "activated_subscription_id": "0d4d24e7-c3d5-4ed6-b4c3-90379af7fdfd",
    "status": "paid",
    "cycle": "month",
    "amount": 1299000,
    "currency": "IDR",
    "provider": "midtrans",
    "provider_order_id": "EA-d55ba687-445e-41d2-b61b-02ea7683e253",
    "provider_transaction_id": "trx-midtrans-id",
    "payment_type": "bank_transfer",
    "transaction_status": "settlement",
    "paid_at": "2026-06-08T01:00:00+08:00",
    "expires_at": "2026-06-09T00:47:47+08:00",
    "created_at": "2026-06-08T00:47:47+08:00",
    "updated_at": "2026-06-08T01:00:02+08:00"
  }
}
```

## 7. List Payment History

Endpoint:

```http
GET /billing/payments
```

Purpose:

- `superadmin`: list payments across all schools.
- `admin_sekolah`: list only the school's own payment history.

Query params:

- `school_id` optional, only for `superadmin`
- `status` optional: `pending`, `paid`, `failed`, `expired`, `cancelled`
- `search` optional: search by school name, plan name, or provider order id
- `page` optional
- `per_page` optional

Success response example:

```json
{
  "success": true,
  "message": "payments retrieved",
  "data": [
    {
      "id": "d55ba687-445e-41d2-b61b-02ea7683e253",
      "school_id": "4c507ee2-dd3e-41d3-830e-a91c09f2b0f2",
      "school_name": "Sekolah QA Subscription",
      "plan_id": "aae1c1c4-3d39-4de7-a41d-deca9644ac76",
      "plan_name": "Pro",
      "created_by_user_id": "e82c418e-eb64-414d-9463-aab0e197b5fc",
      "status": "paid",
      "cycle": "month",
      "amount": 1299000,
      "currency": "IDR",
      "provider": "midtrans",
      "provider_order_id": "EA-d55ba687-445e-41d2-b61b-02ea7683e253",
      "provider_transaction_id": "trx-midtrans-id",
      "payment_type": "bank_transfer",
      "transaction_status": "settlement",
      "paid_at": "2026-06-08T01:00:00+08:00",
      "expires_at": "2026-06-09T00:47:47+08:00",
      "created_at": "2026-06-08T00:47:47+08:00",
      "updated_at": "2026-06-08T01:00:02+08:00"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

## 8. Midtrans Webhook

Endpoint:

```http
POST /billing/webhooks/midtrans
```

Purpose:

- Receive Midtrans notification to finalize and sync payment status.

Notes:

- No bearer token required.
- This endpoint is for Midtrans server-to-server callback, not for frontend use.

## Client Integration Summary

Recommended client flow:

1. Call `GET /schools/plans`
2. Call `GET /schools` or `GET /schools/{id}/subscription`
3. Call `POST /schools/{id}/subscription/checkout`
4. Open `provider_redirect_url`
5. Poll `GET /schools/{id}/subscription/payments/{payment_id}`
6. If payment status is `paid`, refresh `GET /schools/{id}/subscription`
7. Use `GET /billing/payments` for payment history screens

Recommended UI usage:

- Superadmin school list: `GET /schools`
- Superadmin payment monitoring: `GET /billing/payments`
- School admin current subscription: `GET /schools/{id}/subscription`
- School admin payment history: `GET /billing/payments`
