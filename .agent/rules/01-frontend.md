---
trigger: always_on
---

# PART 4: THE AESTHETIC ENGINE (UI/UX MASTERY)

**STANDARD:** v0.dev / LOVABLE.AI / APPLE HUMAN INTERFACE GUIDELINES

## 4.1. VISUAL HIERARCHY AND SPACING PROTOCOLS

You are to assume the role of a Lead Product Designer. "Functionality without Beauty is Failure."

### 4.1.1. THE "BREATHING ROOM" MANDATE

- **Padding:** Default container padding is `p-6` (24px) or `p-8` (32px).
- **Gap:** Default grid gap is `gap-6`.
- **Forbidden:** Cramped layouts (`p-2`, `gap-2`) are strictly prohibited unless building dense data tables.
- **White Space:** Treat white space as a distinct design element. Do not fill every pixel.

### 4.1.2. TYPOGRAPHY SYSTEM (INTER / GEIST SANS)

- **Headings:** Must use `tracking-tight` (-0.025em) and `font-semibold` or `font-bold`.
- **Body:** Must use `text-foreground` (Primary) and `text-muted-foreground` (Secondary).
- **Scale:** Use semantic sizing (`text-xl` for card titles, `text-sm` for metadata).
- **Contrast:** Ensure WCAG AA compliance automatically.

### 4.1.3. GLASSMORPHISM AND DEPTH

- **Surface:** Use `backdrop-blur-md` combined with `bg-background/80` or `bg-white/50` for sticky headers, modals, and overlays.
- **Borders:** Use subtle, translucent borders (`border-white/10` or `border-border/40`). Never use solid black borders (`border-black`).
- **Shadows:** Use `shadow-sm` for interactive cards, `shadow-lg` for dropdowns/modals.

## 4.2. MICRO-INTERACTIONS (THE DELIGHT FACTOR)

Every user action must have a corresponding visual response. "Dead" UI is a system error.

### 4.2.1. TACTILE FEEDBACK LOOP

- **Hover State:** Interactive elements must scale up (`hover:scale-[1.02]`) or brighten (`hover:brightness-110`).
- **Active State:** Buttons must scale down (`active:scale-[0.98]`) to simulate a physical press.
- **Transition:** ALL state changes must use `transition-all duration-200 ease-in-out`.

### 4.2.2. ANIMATION PRIMITIVES

- **Entrance:** Lists and Cards must stagger in using `framer-motion` or `tailwindcss-animate`.
  - *Spec:* `initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}`.
- **Exit:** Elements (Toasts, Modals) must fade out and scale down (`exit={{ opacity: 0, scale: 0.95 }}`).
- **Layout:** Use `layout` prop in Framer Motion to automatically animate layout shifts (reordering lists).

### 4.2.3. LOADING STATE PSYCHOLOGY

- **Forbidden:** Using simple text like "Loading..." or generic browser spinners.
- **Mandatory:** Use **Skeleton Loaders** (`animate-pulse bg-muted rounded-md`) that mimic the exact shape and size of the content being loaded.
- **Optimistic UI:** For mutations (Like, Save, Delete), update the UI *immediately* before the API responds. Rollback on error.

---

# PART 5: FRONTEND ENGINEERING ARCHITECTURE (WEB)

**STACK:** NEXT.JS (APP ROUTER) / REACT / TYPESCRIPT

## 5.1. STATE MANAGEMENT DISCIPLINES

- **Server State:** STRICTLY handled by **TanStack Query (React Query)** or **SWR**. Raw `useEffect` for data fetching is a Grade F failure.
- **Client Global State:** STRICTLY handled by **Zustand**. Context API is reserved for static dependencies (Theme, Auth Session) only.
- **URL State:** Filters, Pagination, and Search Queries MUST be synced to the URL (`searchParams`). This ensures shareability.

## 5.2. DATA FETCHING AND CACHING STRATEGY

- **Server Components (RSC):** Fetch data directly in the component using `async/await`. Pass sanitized data to Client Components.
- **Deduplication:** Leverage Next.js `fetch` caching automatically.
- **Waterfall Prevention:** Use `Promise.all()` for parallel data fetching. Do not await sequentially unless dependent.

## 5.3. FORM HANDLING AND VALIDATION

- **Schema First:** Define the validation schema using **Zod** before writing the form.
- **Integration:** Connect Zod schema to **React Hook Form** via `zodResolver`.
- **UX Pattern:**
  - Validate on Blur (`mode: 'onBlur'`).
  - Show inline error messages in `text-destructive text-sm`.
  - Disable submission button while `isSubmitting` is true.

## 5.4. ERROR BOUNDARIES AND RESILIENCE

- **Component Level:** Wrap complex widgets (Charts, Data Tables) in an `<ErrorBoundary>` to prevent the "White Screen of Death".
- **Global Level:** Create `error.tsx` and `not-found.tsx` in the App Router root.
- **Recoverability:** Error UI must include a "Try Again" button that resets the error boundary or invalidates the query.

---

# PART 16: EXTREME PERFORMANCE ENGINEERING (WEB VITALS / LATENCY)

**STANDARD:** CORE WEB VITALS (GOOGLE) / P99 LATENCY SLO

## 16.1. FRONTEND PERFORMANCE (THE "INSTANT" MANDATE)

- **Core Web Vitals Thresholds:**
  - **LCP (Largest Contentful Paint):** Must be < 2.5s on 4G networks.
  - **INP (Interaction to Next Paint):** Must be < 200ms.
  - **CLS (Cumulative Layout Shift):** Must be < 0.1.
- **Optimization Tactics:**
  - **Image Optimization:** STRICTLY use modern formats (`AVIF`, `WebP`) with explicit `width`/`height` attributes to prevent layout shifts. Use `priority={true}` for LCP images (Hero sections).
  - **Code Splitting:** Implement Route-based splitting (`React.lazy`, `dynamic()`). Keep the initial JS bundle size < 100KB (gzipped).
  - **Font Loading:** Use `font-display: swap` or `optional` to prevent FOIT (Flash of Invisible Text). Self-host fonts to avoid external DNS lookups.

## 16.2. BACKEND PERFORMANCE PROFILING

- **Profiling Standards:**
  - **CPU Profiling:** Use `pprof` (Go) or `py-spy` (Python) to identify "Hot Paths". Optimize loops and regex operations found in these paths.
  - **Memory Leak Detection:** Monitor Heap usage over 24 hours. If usage grows linearly without GC reclamation, trigger a Heap Dump analysis.
- **Database Query Analysis:**
  - **Explain Analyze:** For any query taking > 100ms, run `EXPLAIN ANALYZE` (Postgres) to inspect the Query Plan.
  - **Index Usage:** Verify that `Index Scan` is used instead of `Seq Scan` (Full Table Scan) for large datasets.

## 16.3. CDN AND EDGE COMPUTING

- **Edge Caching:** Cache static assets (JS, CSS, Images) at the Edge (Cloudflare/Vercel/AWS CloudFront) with `Cache-Control: public, max-age=31536000, immutable`.
- **Stale-While-Revalidate:** Use `SWR` strategies for dynamic content that can tolerate slight staleness (e.g., Blog lists, Products) to serve instant responses while updating in the background.
