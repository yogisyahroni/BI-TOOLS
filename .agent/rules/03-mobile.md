---
trigger: always_on
---

# PART 6: MOBILE ENGINEERING EXCELLENCE (NATIVE & CROSS-PLATFORM)

**STACK:** FLUTTER / REACT NATIVE / KOTLIN MULTIPLATFORM

## 6.1. ARCHITECTURAL PATTERNS

- **Modularization:** You must decouple the codebase into Feature Modules.
  - Structure: `:core:network`, `:core:database`, `:feature:auth`, `:feature:dashboard`.
- **Offline-First Mandate:** The Local Database (Room/SqlDelight/WatermelonDB) is the Single Source of Truth. The Network is merely a synchronization mechanism.
- **Sync Engine:** Implement a `WorkManager` (Android) or Background Fetch task to sync data when connectivity returns.

## 6.2. REACT NATIVE SPECIFICS

- **Routing:** Use **Expo Router** (File-based routing) exclusively.
- **Styling:** Use **NativeWind** (Tailwind CSS for RN) for styling consistency with Web.
- **Performance:**
  - Use `FlashList` instead of `FlatList` for large lists (100+ items).
  - Use `Reanimated 3` for all animations (run on UI Thread). Avoid `Animated` API bridge crossings.

## 6.3. FLUTTER SPECIFICS

- **State Management:** Use **Riverpod** with Code Generation (`@riverpod`). Avoid `GetX`.
- **Linting:** Enforce `flutter_lints` and `very_good_analysis` rulesets.
- **Responsiveness:** Use `LayoutBuilder` and `MediaQuery` to support Foldables and Tablets.

## 6.4. MOBILE SECURITY HARDENING

- **Root/Jailbreak Detection:** Implement `flutter_jailbreak_detection` or equivalent. If compromised, wipe sensitive tokens and exit.
- **Certificate Pinning:** Pin the SHA-256 hash of the backend's SSL certificate to prevent MitM (Man-in-the-Middle) attacks.
- **Screenshot Prevention:** Block screenshots on sensitive screens (e.g., OTP, Payment) using `WindowManager.FLAG_SECURE` (Android).
