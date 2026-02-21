---
trigger: always_on
---

# PART 11: SPECIALIZED STACKS (WEB3 & DESKTOP)

**STACK:** SOLIDITY / RUST (TAURI) / ELECTRON / FOUNDRY

## 11.1. WEB3 AND BLOCKCHAIN ENGINEERING

- **Smart Contract Development:**
  - **Toolchain:** Use **Foundry** (`forge`, `cast`, `anvil`) for development, testing, and deployment. Hardhat is legacy.
  - **Standards:** STRICTLY use **OpenZeppelin Contracts** for ERC-20, ERC-721, and AccessControl. Never write token logic from scratch.
  - **Upgradability:** If using Proxies (UUPS/Transparent), you must initialize storage variables correctly to prevent collisions.

## 11.2. SMART CONTRACT SECURITY (GRADE S++)

- **Reentrancy Guard:** Apply `nonReentrant` modifier (OpenZeppelin) to ALL external-facing functions that modify state.
- **Checks-Effects-Interactions:** Follow this pattern religiously.
  1. **Checks:** Validate inputs and conditions.
  2. **Effects:** Update state variables.
  3. **Interactions:** Make external calls (transfer ETH, call other contracts).
- **Oracle Manipulation:** Use **Chainlink Data Feeds** or TWAP (Time-Weighted Average Price). Never rely on `block.timestamp` or spot price from a single DEX.

## 11.3. DESKTOP ENGINEERING (TAURI / ELECTRON)

- **Security Architecture:**
  - **Context Isolation:** MUST be enabled (`contextIsolation: true`).
  - **Sandbox:** MUST be enabled (`sandbox: true`).
  - **Node Integration:** MUST be disabled (`nodeIntegration: false`) in Renderers.
- **IPC (Inter-Process Communication):**
  - **Scope:** Whitelist allowed backend commands explicitly.
  - **Validation:** Validate all IPC payloads using Zod/Pydantic before execution.
  - **Wildcards:** Deny all `*` wildcards in IPC handlers.

## 11.4. DESKTOP DISTRIBUTION AND SIGNING

- **Code Signing (Windows):** Warn the user that an EV/OV Code Signing Certificate is required to bypass SmartScreen filters.
- **Notarization (macOS):** Implement the `xcrun notarytool` workflow to pass Apple Gatekeeper.
- **Auto-Update:** Implement strict signature verification for update binaries (Tauri Updater / electron-updater).
