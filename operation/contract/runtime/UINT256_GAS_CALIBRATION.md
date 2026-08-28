# Uint256 Gas Calibration

Measurement baseline: `8f6196d0ed13cc065114be156d3083caee8495a0`.

Environment: Go 1.25.9, Gno `v0.0.0-20260513121932-eb9f51920519`, darwin/arm64, Apple M1 Max. Wall-clock benchmark values are reference data, not consensus pass criteria.

## Boundary

Fixed stdlib and pure-package initialization occurs outside invocation gas. Contract and host package loading, snapshot restore, typed invocation, query conversion, and snapshot capture use the execution meter. Tests read package-private meter and allocator status; no public metrics API is added.

## Selected Flat Gas

| Native | Previous provisional | Selected |
| --- | ---: | ---: |
| `_nativeCanonicalToHex` | 10,000 | 10,000 |
| `_nativeMulDiv` | 25,000 | 40,000 |
| `_nativeSqrt` | 15,000 | 20,000 |

All inputs and outputs are bounded to uint256 canonical representations. Canonical conversion is the 10,000 base tier. Worst observed host MulDiv CPU was about 3.7 times canonical conversion and allocated about 4.5 times as much, rounded to 40,000. Sqrt CPU was about 1.7 times and allocation about 2.4 times the base, rounded to 20,000. These rounded tiers provide a conservative margin while leaving full-range decimal-native query operations below the fixed 1,000,000 gas limit. Schedule identity is `mitum-u256-gas-v1`.

## Measured Workloads

Under the optimized wrapper, maximum canonical validation consumed about 362k gas and maximum canonical-to-Uint conversion 765,291 gas. Maximum decimal-native MulDiv measured 399,058 write gas and 398,343 query gas; Sqrt measured 377,590 write gas and 376,995 query gas. Query allocator usage was 3,945,063 bytes for MulDiv and 3,944,752 for Sqrt, leaving 249,241 and 249,552 bytes, about 6%, headroom. Three repeated measurements must produce identical gas and allocator readings.

Maximum Uint-based MulDiv remains unsupported at the 5,000,000 write limit because three Uint-to-decimal formats dominate the path; its elevated-limit measurement was about 6.4M gas. Use `MulDivCanonicalDecimal` for full-range persistent-state calculations. Maximum Uint Sqrt remains supported.

Query allocation headroom is narrow and is an RC risk. A single maximum decimal-native operation is required to pass repeatedly. Composing multiple operations in one query is not guaranteed and must fail deterministically at the current limit rather than increasing it.

## Consensus Identity

- Public package: `mitum/math/v1/u256`
- Wrapper SHA-256: `43980a468b2a107b64c3efa4ced7568b5f21a8f012dbcdd22fd94dae03cee899`
- Onbloc bundle SHA-256: `2e60329f269f003bc64c023aeb4f3288ca597e2cf906819b37f46a6de0d7aeb5`
- Canonical policy: `mitum-u256-canonical-decimal-v1`
- Native semantics: `mitum-u256-native-v2`
- Gas schedule: `mitum-u256-gas-v1`

## Host Benchmarks

Run:

```sh
go test ./operation/contract/runtime -run '^$' -bench 'Uint256' -benchmem -count=5
```

Five-count measurements observed canonical-to-hex at 1.014 us, 568 B and 10 allocations; canonical parsing at 0.898 us, 344 B and 7 allocations; MulDiv at 3.242 us, 1,544 B and 27 allocations; and Sqrt at 1.449 us, 816 B and 16 allocations. Re-run on the release toolchain and record worst observations; wall time does not alter consensus gas.

## Reproduction

```sh
go test ./operation/contract/runtime -run 'TestUint256FinalCalibration' -count=1 -v
go test ./operation/contract/runtime/...
go test -race -timeout 10m ./operation/contract/runtime -run '^TestUint256FinalCalibration'
```

Focused race verification completed with U6A-R2A at 3/3 commands passing, U6A-R2B at 4/4, and U6A-R2C at 4/4. It produced no race findings and no focused timeouts. Each run emitted the known macOS malformed `LC_DYSYMTAB` linker warning.

The package-wide runtime race command timed out at both 10 and 20 minutes without emitting a race detector finding; the 20-minute run was executing `TestUint256OptimizedMaxWriteAndQuery` when it timed out. This package-wide timeout is deferred as legacy audit work and does not block the U6A-R commit. Pure arithmetic differential and non-uint256 runtime race coverage move to U6A-LR.

## Nested Shared Gas

C2 nested Uint256 integration is verified by `TestNestedUint256ExactSharedGasMeterIdentity`, which observes the top-level machine and two nested machines receiving the exact same gas meter instance. All three machines receive the unchanged write allocator limit. Allocators remain machine-local; aggregate allocation protection continues to rely on depth, touched-contract, and shared-gas limits rather than a new shared allocator.

`TestNestedUint256NativePathsAndOverlayVisibility` exercises canonical conversion, MulDiv, and Sqrt through the public `mitum/math/v1/u256` API. Its second nested call reads the first call's session-overlay snapshot before top-level merges are applied. `TestNestedUint256GasAccumulationDeterministic` measured 481,425 gas for one nested workload and 866,716 for the multi-nested workload, a deterministic 385,291 delta on the shared meter.

`TestNestedUint256OutOfGasAtomicRollback` consumed 5,026,357 against the fixed 5,000,000 limit and returned no merges while preserving caller and both target snapshots. `TestNestedUint256NativeFailureAtomicRollback` likewise preserved all three snapshots and returned the stable denominator-zero reason with no merges. Caller/Origin ABI was applied and verified in C3. The model-owned currency account/balance overlay was applied and verified in C4, including shared-gas snapshot/currency atomic rollback. The U6A-LR legacy race audit moves to UF2.

Write/query gas and allocation limits, package setup accounting, snapshot codec, and digest schema are unchanged.
