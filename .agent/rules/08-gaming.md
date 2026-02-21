---
trigger: always_on
---

# PART 22: GAME ENGINEERING & HIGH-PERFORMANCE COMPUTING

**STANDARD:** ENTITY-COMPONENT-SYSTEM (ECS) / DATA-ORIENTED DESIGN

## 22.1. MEMORY MANAGEMENT & GARBAGE COLLECTION

- **Object Pooling:** STRICTLY FORBIDDEN to instantiate/destroy objects (bullets, enemies) inside the Game Loop (`Update()`).
  - *Mandate:* Use pre-allocated Object Pools. Reuse entities to prevent GC Spikes and frame drops.
- **Data Locality:**
  - Use **Structs** over Classes (C#) or POD types (C++) to ensure cache coherence.
  - Process contiguous arrays of components (Data-Oriented Design) rather than chasing pointers.

## 22.2. GAME LOOP ARCHITECTURE (UNITY / UNREAL / BEVY)

- **Tick Rate Decoupling:**
  - **Physics:** Run on a fixed timestep (`FixedUpdate`, e.g., 50Hz) for deterministic simulation.
  - **Rendering:** Run on variable timestep (`Update`) with interpolation for smooth visuals.
- **ECS Pattern (Entity-Component-System):**
  - **Entities:** Just IDs.
  - **Components:** Pure Data (no logic).
  - **Systems:** Logic that iterates over components.
  - *Rule:* Never mix logic inside Component classes. Keep data and behavior separate.

## 22.3. SHADER & GPU OPTIMIZATION

- **Draw Calls:** Minimize draw calls by using **GPU Instancing** and **Texture Atlases**.
- **Overdraw:** Render opaque objects front-to-back. Render transparent objects back-to-front.
- **Material Complexity:** Bake lighting into Lightmaps for static geometry. Avoid real-time global illumination on mobile targets.
