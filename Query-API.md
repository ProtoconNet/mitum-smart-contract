# Query API

이 문서는 현재 저장소에서 지원하는 contract query HTTP API 사용법을 정리한다.

## Endpoint

- `POST /contract/{contract}/query`

## Request Body

request body는 JSON object이고, 반드시 `function` 키를 포함해야 한다.

```json
{
  "function": "GetConfig"
}
```

현재 query endpoint는 raw HTTP body size limit도 적용한다.

- max query body size: `128 KiB`

### Scalar Arg Rule

scalar query arg는 기존처럼 plain string이다.

지원 scalar:

- `string`
- `bool`
- `int`
- `int64`
- `uint64`

예:

```json
{
  "function": "GetUser",
  "name": "alice"
}
```

### Request Payload Limits

decode 후 `map[string]string` payload에도 size limit이 적용된다.

- max entries: `64`
- max key bytes: `128`
- max value bytes: `16 KiB`
- max total key+value bytes: `64 KiB`

reserved key인 `function` 도 일반 entry와 동일하게 계산된다.

`_sender` 는 더 이상 query execution context sender 로 해석되지 않는다. 요청에 들어와도 일반 unused key처럼 payload limit 계산에 포함될 뿐이며, query contract는 `chain.QueryContext`에서 sender에 접근할 수 없다.

### Query Height Semantics

query contract에서 `ctx.GetHeight()`는 현재 query가 읽는 state/view height다. snapshot-backed typed Gno contract에서는 snapshot state height를 반환한다.

현재 chain head height가 필요하면 query 함수에서 `ctx.GetCurrentHeight()`를 사용한다. digest query path는 database의 current last block height를 runtime에 전달하며, 이 값은 state/view height와 다를 수 있다. write/register/call path에서는 current head height ABI가 없으며, write execution/block height는 `ctx.GetHeight()`로 읽는다. `chain.CurrentHeight()`는 더 이상 canonical contract surface가 아니다.

write-only `ctx.GetBlockTime()`은 proposal inclusion timestamp를 위한 API이며 query context에는 제공되지 않는다.

### Balance Lookup Native

contract query/write 함수는 `mitum/chain` host native로 현재 balance state를 조회할 수 있다.

```go
amount, ok := chain.BalanceOf(addr, currency)
```

Semantics:

- `amount`는 decimal amount string이다.
- account, currency design, balance state가 모두 있으면 `ok == true`.
- account 없음, currency 없음, 해당 currency balance state 없음, malformed address, malformed currency는 `("", false)`로 반환된다.
- zero balance는 `"0", true`로 반환되어 not found와 구분된다.
- state decode/type mismatch는 state corruption 성격의 host native failure이며, raw internal detail은 panic sanitization 정책에 따라 HTTP response에 직접 노출되지 않는다.

### SHA3-256 Native

contract query/write 함수는 deterministic pure host native로 SHA3-256 digest를 계산할 수 있다.

```go
digest := chain.SHA3Sum256(data)
```

Semantics:

- input은 `[]byte(data)` raw byte sequence다.
- hex decode, numeric parse, UTF-8 text normalization은 하지 않는다.
- output은 lowercase hex digest string이다.
- failure result는 없다.
- `"ff"`는 bytes `{0x66, 0x66}`로 해시된다.

## Query Result Shape

현재 query result는 다음 형태를 지원한다.

- `T`
- `(T, bool)`

여기서 `T`는 아래 범위까지 지원된다.

- scalar
- named struct
- `map[string]scalar`
- `map[string]named-struct`
- `[]scalar`
- `[]named-struct`

JSON/HAL 응답에서의 shape:

- scalar -> JSON scalar
- struct -> JSON object
- map -> JSON object
- slice -> JSON array

HTTP 응답에서는 query 함수의 반환값을 항상 `_embedded.output` 아래에 둔다.

- single-result query: `output.result`만 포함
- `(T, bool)` query: `output.result`와 `output.ok` 포함

`_embedded.result`와 `_embedded.ok` 형태는 더 이상 제공하지 않는다. 이는 메타데이터와 함수 output을 분리하기 위한 즉시 적용 API cleanup이며 compatibility shim은 없다.

## HAL Response

응답은 HAL JSON이다.

예:

```json
{
  "_embedded": {
    "contract": "mitum....",
    "function": "GetConfig",
    "engine": "gno-snapshot-v1",
    "read_only": true,
    "output": {
      "result": {
        "Owner": "alice"
      }
    }
  },
  "_links": {
    "self": { "href": "/contract/mitum..../query" },
    "design": { "href": "/contract/mitum...." },
    "block": { "href": "/block/123" }
  }
}
```

## `(T, bool)` Semantics

예:

```json
{
  "_embedded": {
    "contract": "mitum....",
    "function": "GetValueIfPresent",
    "engine": "gno-snapshot-v1",
    "read_only": true,
    "output": {
      "result": "5",
      "ok": true
    }
  }
}
```

- `output.ok == true`
    - `output.result`는 실제 조회된 값을 담는다.
- `output.ok == false`
    - `output.result`는 해당 타입의 zero-like JSON shape를 담을 수 있다.
    - 예:
        - struct -> field별 zero value / nil map / nil slice
        - scalar -> zero scalar

## Nil / Empty Result Semantics

### Map

- nil map -> JSON `null`
- empty map -> JSON `{}`

### Slice

- nil slice -> JSON `null`
- empty slice -> JSON `[]`

## Unsupported 범위

현재 아래는 지원하지 않는다.

- anonymous struct arg / result
- non-string map key
- map elem map
- map elem slice
- slice elem map
- slice elem slice
- recursive / mutually recursive struct
- struct query arg
- map query arg
- slice query arg

## Error Cases

대표적인 실패 유형:

- malformed JSON -> `400 Bad Request`
- oversized raw query body -> `413 Request Entity Too Large`
- decoded `callData` payload limit 초과 -> `400 Bad Request`
- missing `function` -> `400 Bad Request`

## curl Examples

scalar query:

```bash
curl -X POST \
  -H 'Content-Type: application/json' \
  -d '{"function":"GetOwner"}' \
  http://localhost:54320/contract/<contract>/query
```

scalar arg query:

```bash
curl -X POST \
  -H 'Content-Type: application/json' \
  -d '{"function":"GetUser","name":"alice"}' \
  http://localhost:54320/contract/<contract>/query
```

scalar multi-arg query:

```bash
curl -X POST \
  -H 'Content-Type: application/json' \
  -d '{"function":"GetUserTagAt","name":"owner","index":"0"}' \
  http://localhost:54320/contract/<contract>/query
```
