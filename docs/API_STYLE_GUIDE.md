# API Style Guide

## 1. Governance & Standards

All APIs must adhere to **OpenAPI 3.1** standards. The single source of truth is `backend/docs/openapi.yaml`.

## 2. Naming Conventions

### 2.1 URIs

- **Kebab-case** for path segments: `/api/user-profiles` (NOT `/api/userProfiles`).
- **Plural nouns** for resources: `/api/users` (NOT `/api/user`).
- **Lowercase** always.

### 2.2 Fields (JSON)

- **CamelCase** for properties: `firstName`, `createdAt`.
- **No Hungarian Notation**: `id` (NOT `userId` inside a User object).

### 2.3 Query Parameters

- **CamelCase**: `?page=1&pageSize=20`.
- **Standard Pagination**: Always use `page` and `limit`.

## 3. HTTP Methods

- `GET`: Retrieve resources. Safe & Idempotent.
- `POST`: Create resources. Not Idempotent.
- `PUT`: Full update/replace. Idempotent.
- `PATCH`: Partial update. Not strictly Idempotent (but should be).
- `DELETE`: Remove resources. Idempotent.

## 4. Status Codes

- `200 OK`: Standard success.
- `201 Created`: Resource created (include `Location` header).
- `204 No Content`: Successful action with no body (DELETE).
- `400 Bad Request`: Validation failure.
- `401 Unauthorized`: Missing/invalid token.
- `403 Forbidden`: Valid token, but insufficient permissions.
- `404 Not Found`: Resource does not exist.
- `409 Conflict`: Duplicate resource or state conflict.
- `500 Internal Server Error`: Server blew up.

## 5. Error Responses

All errors must return the standard `ErrorResponse` schema:

```json
{
  "error": "Human readable message",
  "code": "MACHINE_READABLE_CODE",
  "details": { "field": "reason" }
}
```

## 6. Security

- **Bearer Auth**: JWT in `Authorization` header.
- **HSTS**: Enforced on all responses.
- **CSP**: Enforced via headers.

## 7. Versioning

- **URI Versioning**: `/api/v1/...`
- **Breaking Changes**: Requires incrementing version.

## 8. Documentation (OpenAPI)

- **Summary**: Required. Concise (< 50 chars).
- **Description**: Required. Markdown supported.
- **OperationId**: Required. VerbResource format (`getUser`, `createUser`).
- **Tags**: Exactly one tag per operation.
- **Responses**: Must document `200`, `400`, `401`, `403`, `500` at minimum.
