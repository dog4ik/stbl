#!/usr/bin/env bash

# curl -X POST "http://localhost:4000/api/v1/payouts" \
#   -H 'content-type: application/json' \
#   -H 'authorization: Bearer d72657b9c6c5863c1a6e' \
#   -d '
# {
#   "amount": 45000,
#   "currency": "ARS",
#   "bank_account": {
#     "account_number": "1234567890123456789012"
#   },
#   "customer": {
#     "email": "test@gmail.com",
#     "ip": "127.0.0.1"
#   },
#   "order_number": "5113262000005644"
# }
# '

curl -X POST "http://localhost:4000/api/v1/payments" \
  -H 'content-type: application/json' \
  -H 'authorization: Bearer d72657b9c6c5863c1a6e' \
  -d '
{
  "amount": 4500000,
  "currency": "ARS",
  "product": "test product",
  "customer": {
    "email": "test@gmail.com",
    "ip": "127.0.0.1"
  },
  "order_number": "5113262000005644"
}
'
