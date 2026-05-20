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

reserved key인 `function` 과 `_sender` 도 일반 entry와 동일하게 계산된다.

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

`(T, bool)` query는 기존처럼 `result`와 `ok`를 함께 반환한다.

## HAL Response

응답은 HAL JSON이다.

예:

```json
{
  "_embedded": {
    "contract": "mitum....",
    "function": "GetConfig",
    "engine": "gno-snapshot-v1",
    "result": {
      "Owner": "alice"
    },
    "read_only": true
  },
  "_links": {
    "self": { "href": "/contract/mitum..../query" },
    "design": { "href": "/contract/mitum...." },
    "block": { "href": "/block/123" }
  }
}
```

## `(T, bool)` Semantics

- `ok == true`
    - `result`는 실제 조회된 값을 담는다.
- `ok == false`
    - `result`는 해당 타입의 zero-like JSON shape를 담을 수 있다.
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
