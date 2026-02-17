# 3. Formula Engine Design

Date: 2026-02-17

## Status

Accepted

## Context

Users need to perform calculations on datasets (e.g., `[Sales] * [Quantity]`) and visualizations without altering the source data schema. We need a DAX-like formula engine that runs efficiently on the backend.

## Decision

We implemented a custom Formula Engine in Go with the following components:

1. **Lexer/Parser:** Custom recursive descent parser generating an AST from Excel-like syntax.
2. **Evaluator:** A `BatchEvaluator` that processes entire columns (arrays) at once rather than row-by-row, leveraging Go's slice performance.
3. **Persistence:** Calculated fields are stored as JSONB in the `DashboardCard` model rather than physically altering data tables. This allows for dynamic, non-destructive calculations.

## Consequences

- **Performance:** Batch processing minimizes function call overhead.
- **Flexibility:** Users can define ad-hoc formulas without DBA intervention.
- **Complexity:** We maintain our own parser and function registry, which requires ongoing maintenance as new functions are requested.
