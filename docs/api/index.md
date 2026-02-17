# API Documentation Portal

InsightEngine AI provides a comprehensive REST API for automation and integration.

## OpenAPI Specification

The full API specification is available via Swagger UI.

[**Launch Swagger UI**](/api/docs)

*(Note: Ensure the backend server is running to access the interactive documentation)*

## Authentication

All API requests require a Bearer Token.

```bash
Authorization: Bearer <your_api_token>
```

Generate tokens in **Settings > API Access**.

## Common Endpoints

- `GET /api/v1/dashboards`: List dashboards
- `POST /api/v1/queries/run`: Execute a SQL query
- `GET /api/v1/datasets`: specific dataset metadata
