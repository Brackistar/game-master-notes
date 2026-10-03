# TCL Tab 11 Gen 2 Acceptance Profile

## Metadata

- Created: 2026-10-02
- Last updated: 2026-10-02
- User: Brackistar
- File: `docs/2026-10-02-tcl-tab-11-gen-2-acceptance.md`
- Status: automated harness ready; device measurements pending connection

## Reference device

- Device: TCL Tab 11 Gen 2
- RAM: 6 GB
- Android: 15 / API 35
- SoC: MediaTek Helio G80
- Runtime policy: ARM64 CPU for MiniLM and llama.cpp; no accelerator assumption

The device supersedes the earlier generic low-end reference for POC acceptance. It remains a single-device gate, so results establish compatibility for this tablet rather than broad Android performance coverage.

## Automated encoder checks

Connect the tablet with USB debugging enabled and confirm it appears under `adb devices -l`. From `android/`, run:

```powershell
.\gradlew.bat :core:retrieval:connectedDebugAndroidTest
adb logcat -d -s GmnTabletAcceptance:I
```

The instrumented lane verifies the Android ONNX vector against a pack-builder golden vector and records:

- first-call encoder latency, including asset verification, copy, and session load;
- five warm-query latencies;
- total and native PSS;
- manufacturer, model, and Android API level.

## POC gates

Record results before changing thresholds. Initial investigation thresholds for this device are:

| Measurement | Investigation threshold |
| --- | ---: |
| Warm query encoding | 500 ms |
| Hybrid retrieval after warm-up | 1,000 ms |
| Cancellation acknowledgement | 500 ms |
| Encoder plus vector-index PSS increase | 250 MB |
| Recall@4 overall | 85% |
| Recall@4 paraphrase/relationship | 80% |
| nDCG@4 | 0.80 |

Cold encoder initialization is recorded separately and is not charged to the generation deadline. LLM acceptance still requires the selected GGUF files to be installed on the tablet.

## Manual observation boundary

The automated harness captures timings and memory. Thermal throttling, UI responsiveness, answer usefulness, and whether cancellation feels immediate still require observation on this one physical tablet. Those results must not be inferred from emulator or JVM checks.
