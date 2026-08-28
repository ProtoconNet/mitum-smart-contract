# Release Candidate Baseline

UF1 freezes the Uint256 hybrid consensus baseline. The UF1 commit is the commit containing this document and `operation/contract/runtime/UINT256_CONSENSUS_BUNDLE.json`.

## Commit Lineage

- `39a479ee` classifier baseline
- `2fc83d1a` U1 source pinning
- `f2b51413` U2 package loader
- `a31739f1` U3 arithmetic semantics
- `b2e90ad9` U4 native math API
- `87e34da7` U5 canonical boundary
- `8f6196d0` U5O-B optimized canonical path
- `4bfa3b20` U6A-R calibrated resources
- `de494c95` C1 contract runtime session
- `d4d7ef5a` C2 nested Uint256 shared gas
- `d6bb2d29` C3 Caller/Origin ABI
- `ccbcdd8e` C4 currency overlay

## Local Replace Topology

- `github.com/imfact-labs/mitum2 => /Users/soonkukkang/go/src/github.com/imfact-labs/mitum2-serial-process` at `62322414b54d0575f6eac2b2258e7288210e62e9`
- `github.com/imfact-labs/currency-model => /Users/soonkukkang/go/src/github.com/imfact-labs/currency-model-serial-process` at `760787752098a1733eaadf0408ec9ba7b9ace26c`

These absolute replacements are the local RC topology. Reproduction requires preparing the same dependency HEADs and recreating equivalent paths or replacements.

## Included Scope

The baseline includes deterministic Onbloc source pinning, exact pure-package import policy, the versioned `mitum/math/v1/u256` wrapper, basic Uint256 semantics, native MulDiv/Sqrt, canonical state/query boundaries, calibrated gas identity, contract-to-contract shared gas, snapshot overlay atomicity, Caller/Origin ABI, top-level currency account/balance overlay, and classifier plus Mitum2 scheduler-hook integration.

## Mandatory MVP Limitations

- Same-block RegisterContract to CallContract visibility is unsupported; RegisterContract cannot make nested calls and nested Initialize is rejected.
- Self-call and reentrancy are rejected. Session overlays are visible only inside one top-level operation.
- Normal operations cannot use same-block contract credit, and contract operations do not see another operation's same-block currency changes.
- Contract runtime adds no broader currency debit. Transfers are allowed only from the first top-level contract; nested transfers are rejected.
- QueryContext has no GetCaller. Authorization remains the contract author's responsibility.
- No additional Mitum2 core change is required for MVP.
- Maximum Uint-based MulDiv is unsupported under the 5M write limit; full-range persistent calculation uses `MulDivCanonicalDecimal`.
- Query allocation headroom is about 249 KB (6%). Allocators remain machine-local.
- Direct Onbloc imports are prohibited. Absolute replace topology must be recreated for deployment.

## Known Issues

- UF1-R verification used `go version go1.25.9 darwin/arm64` with `GOTOOLCHAIN=auto`; both replacement modules declare Go 1.24.0 and no toolchain or module directive was changed.
- U6A and C2/C3/C4 focused race suites passed with zero findings.
- Package-wide runtime race timed out at 10 and 20 minutes and is deferred to UF2 legacy audit.
- macOS malformed `LC_DYSYMTAB` warnings are nonblocking.
- Smart-contract-model consensus, runtime, digest, classifier, and `go test ./...` pass. `go vet ./...` reports four pre-existing self-assignments at `schema_ruleset.go:307`, `:310`, `:314`, and `:315`; the same statements exist at baseline `39a479ee`. They are nonblocking for this scoped RC, and UF1 includes no production cleanup.
- Currency-model's relevant processor passes with `-vet=off`. Its default processor command fails on non-constant format-string vet diagnostics. Its full `-vet=off` suite retains pre-existing failures in `digest.TestLoadManifestPreservesFeeAmounts`, `TestCurrencyOperationReceiptRoundTrip`, and `TestFixedItemDataSizeExecutionFeeReceiptJSONOmitsEmptyExecutionFields`. The dependency HEAD remained unchanged throughout U1-C4. These failures are not attributed to the smart-contract bridge and are nonblocking for scoped validation, but a currency-model full-suite-green claim is unavailable.
- Mitum2 test helpers require the `test` build tag. The tagged scheduler command encountered only two non-constant format-string vet diagnostics under Go 1.25.9; the permitted `-vet=off` fallback passed. Mitum2 full suite is not part of UF1 and no full-suite-green claim is made.
- Narrow query allocation headroom is an RC risk but not a current blocker. Machine-local allocation is a documented limitation.
