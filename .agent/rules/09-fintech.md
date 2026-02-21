---
trigger: always_on
---

# PART 23: FINANCIAL SYSTEMS ENGINEERING (FINTECH)

**STANDARD:** ISO 8583 / DOUBLE-ENTRY BOOKKEEPING / PCI-DSS

## 23.1. NUMERICAL PRECISION (THE "NO FLOATS" LAW)

- **The Cardinal Sin:** NEVER use `float` or `double` for monetary values. Floating point math (`0.1 + 0.2 != 0.3`) causes money to vanish.
- **Mandatory:** Use **Arbitrary-Precision Decimals** (`BigDecimal` in Java, `decimal` in Python/C#, `Shopify/decimal` in Go).
- **Storage:** Store money in the database as **Integers** (cents/micros) or **Decimal(19,4)**.

## 23.2. LEDGER ARCHITECTURE (DOUBLE-ENTRY)

- **Immutability:** Ledger entries are Append-Only. You never `UPDATE` a transaction balance. You insert a correcting entry.
- **The Equation:** `Assets = Liabilities + Equity`. Every transaction must have at least two splits (Debit/Credit) that sum to zero.
- **Idempotency Keys:** Every financial transaction API request MUST contain an `Idempotency-Key` header.
  - *Logic:* If the client retries a timeout, the server returns the *cached result* of the original request instead of charging the card twice.

## 23.3. PAYMENT SWITCH STANDARDS (ISO 8583)

- **Message Packing:** Efficiently pack bitmaps and fields. Do not send JSON to a Payment Switch/HSM unless wrapped.
- **Encryption:** PIN blocks must be encrypted using **3DES/AES** under a Zone Master Key (ZMK). Never log raw PIN blocks or CVV codes.
