---
trigger: always_on
---

# PART 24: EMBEDDED SYSTEMS & IOT (RUST / C / C++)

**STANDARD:** MISRA C / RTOS CONSTRAINTS

## 24.1. SAFETY CRITICAL C/C++

- **Memory Safety:**
  - **Prohibited:** `malloc`/`free` after initialization phase. Use Static Allocation to prevent fragmentation.
  - **Prohibited:** Recursion (Risk of Stack Overflow).
  - **Mandatory:** Check return values of ALL hardware HAL functions.
- **Rust Embedded:**
  - Use `#![no_std]` for bare-metal targets.
  - Use `unwrap()` ONLY during initialization. In the main loop, handle `Result` explicitly.

## 24.2. IOT COMMUNICATION PROTOCOLS

- **MQTT:**
  - **QoS (Quality of Service):** Use QoS 1 (At Least Once) for critical telemetry. QoS 0 is for disposable data only.
  - **Last Will & Testament (LWT):** Configure LWT to notify the broker if the device disconnects ungracefully (power loss).
- **OTA (Over-the-Air) Updates:**
  - **A/B Partitioning:** Always update to a passive partition (Slot B). Verify checksum/signature. Reboot. If boot fails, Watchdog Timer (WDT) must rollback to Slot A automatically.

## 24.3. POWER MANAGEMENT

- **Sleep Modes:** The device must enter Deep Sleep whenever the radio/sensor is idle.
- **Interrupts:** Use GPIO Interrupts instead of Polling loops to wake the CPU.
