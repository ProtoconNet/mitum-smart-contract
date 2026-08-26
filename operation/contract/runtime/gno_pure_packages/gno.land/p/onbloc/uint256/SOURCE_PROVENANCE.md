# Onbloc Uint256 Source Provenance

- Upstream repository: https://github.com/gnolang/gno
- Module path: `github.com/gnolang/gno`
- Module version: `v0.0.0-20260513121932-eb9f51920519`
- Package identity: `gno.land/p/onbloc/uint256`
- Upstream source directory: `examples/gno.land/p/onbloc/uint256`
- Local source directory: `operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256`
- License: BSD-3-Clause; the unmodified package license is preserved as `LICENSE` in this directory.
- Upstream API background: `examples/gno.land/p/onbloc/uint256/README.md` at the pinned module version.
- Local patches to production sources: none.

## Production Bundle

The production bundle contains only these files, in byte-wise lexicographic order:

1. `arithmetic.gno`
2. `bits_table.gno`
3. `bitwise.gno`
4. `cmp.gno`
5. `conversion.gno`
6. `error.gno`
7. `mod.gno`
8. `uint256.gno`
9. `utils.gno`

`SOURCE_MANIFEST.json` records the SHA-256 of every production file and the
canonical bundle SHA-256. The manifest, this provenance document, and LICENSE
are metadata and are not part of the production bundle hash.

## Excluded Upstream Files

- `arithmetic_test.gno`, `bitwise_test.gno`, `cmp_test.gno`,
  `conversion_test.gno`, and `uint256_test.gno`: upstream tests, excluded from
  runtime production source.
- `gnomod.toml`: upstream module metadata, excluded from runtime production
  source.
- `README.md`: upstream documentation, excluded from runtime production source;
  its pinned location is recorded above for attribution and API background.
- Any other test or filetest material: excluded from runtime production source.

## Canonical Bundle Hash

Sort package-root-relative production file paths by their raw UTF-8 bytes.
Paths use `/` as the separator. For each file, append these fields without
padding: the path byte length as an unsigned 64-bit big-endian integer, the
path bytes, the raw file-content byte length as an unsigned 64-bit big-endian
integer, and the raw file-content bytes. The bundle hash is SHA-256 over the
concatenation of all frames. No newline normalization or locale-sensitive
sorting is performed.
