# Contract Language Support

## 목적

이 문서는 현재 프로젝트 기준으로 스마트 컨트랙트 안에서 어떤 Go/Gno 형태가 지원되고, 어떤 형태가 지원되지 않는지 정리한다.

중요한 점은 이 문서가 "일반적인 Go 전체 문법"을 설명하는 것이 아니라, **현재 이 프로젝트의 typed Gno contract runtime과 ABI 정책 기준으로 실제 사용 가능한 범위**를 설명한다는 것이다.

---

## 기본 전제

- 모든 컨트랙트는 typed Gno contract이다.
- 컨트랙트 패키지는 `package contract` 여야 한다.
- 입력 ABI는 단순화되어 있으며, 상태 저장과 query 반환은 그보다 넓은 타입을 지원한다.

---

## 필수 구조

컨트랙트는 최소한 아래 구조를 따라야 한다.

- `package contract`
- `Initialize(ctx chain.WriteContext, ...scalar) error`

`Initialize`는 필수이며, typed Gno contract의 진입 함수다.

---

## 지원되는 입력 타입

현재 입력 ABI는 **write 함수와 query 함수 모두 scalar-only** 이다.

지원되는 scalar 타입:

- `string`
- `bool`
- `int`
- `int64`
- `uint64`

즉, 현재는 아래와 같은 함수 형태가 지원된다.

- write 함수
  - `func Claim(ctx chain.WriteContext) error`
  - `func X(ctx chain.WriteContext, a string, b int64) error`
- query 함수
  - `func X(ctx chain.QueryContext, name string) bool`
  - `func X(ctx chain.QueryContext, index int) (string, bool)`

입력으로 아래 타입은 직접 받을 수 없다.

- struct
- map
- slice

즉, struct/map/slice를 인자로 받는 write/query 함수는 지원되지 않는다.

---

## write 함수와 query 함수의 구분

현재 프로젝트에서 write 함수와 query 함수는 **이름이 아니라 함수 시그니처**로 구분한다.

즉 `Get...`, `Set...`, `Update...` 같은 이름 규칙은 본질이 아니고, runtime/schema는 함수의 형태를 보고 판단한다.

### write 함수

write 함수는 아래 형태를 따른다.

```go
func X(ctx chain.WriteContext, ...scalar) error
```

조건:

- exported 함수
- 첫 번째 인자가 `chain.WriteContext`
- 함수 이름이 `Initialize`가 아님
- 반환값이 **오직 `error` 1개**

즉 반환형이 `error` 하나이면 write 함수로 분류된다.

중요:

- scalar 인자는 0개여도 된다
- 즉 `func Claim(ctx chain.WriteContext) error` 같은 함수도 유효한 write 함수다
- 이전에는 일부 구현이 `ctx` 외 인자 1개 이상을 요구했지만, 현재 정책과 구현은 zero-arg write를 허용한다

### query 함수

query 함수는 아래 형태를 따른다.

```go
func X(ctx chain.QueryContext, ...scalar) T
func X(ctx chain.QueryContext, ...scalar) (T, bool)
```

조건:

- exported 함수
- 첫 번째 인자가 `chain.QueryContext`
- 함수 이름이 `Initialize`가 아님
- 반환값이
  - 1개이거나
  - 2개인데 두 번째가 `bool`

즉 반환형이 `T` 또는 `(T, bool)`이면 query 함수로 분류된다.

중요:

- 반환값이 `error` 하나인 함수는 query가 아니다
- 즉 `func Claim(ctx chain.WriteContext) error`는 query로 해석되면 안 되고 write로 분류되어야 한다
- `(T, bool)`의 두 번째 반환값은 반드시 `bool`이어야 한다

### Initialize는 예외

`Initialize`는 write/query 일반 규칙과 별도로 취급되는 특수 함수다.

형태:

```go
func Initialize(ctx chain.WriteContext, ...scalar) error
```

이 함수는 컨트랙트 초기화 전용 entrypoint이며, 일반 write 함수와는 별개로 처리된다.

중요:

- `Initialize`는 scalar arg를 받을 수 있다
- 전달은 순서 기반이 아니라 **이름 기반**이다
- register payload의 `init_data` key가 `Initialize` 파라미터 이름과 일치해야 한다

예:

```go
func Initialize(ctx chain.WriteContext, owner string, label string, limit int64) error
```

register payload:

```json
{
  "init_data": {
    "owner": "alice",
    "label": "demo",
    "limit": "10"
  }
}
```

정책:

- `init_data["owner"]` -> `owner`
- `init_data["label"]` -> `label`
- `init_data["limit"]` -> `limit`
- required key 누락 시 실패
- unknown key 존재 시 실패
- scalar parse 실패 시 실패

추가로 `init_data` 자체에도 payload limit이 있다.

- 최대 entry 수: `64`
- key 최대 길이: `128 bytes`
- value 최대 길이: `16 KiB`
- 전체 key+value 총합: `64 KiB`

### 요약

- `... -> error` 이면 write
- `ctx`만 받고 `error`를 반환하는 함수도 write
- `... -> T` 또는 `... -> (T, bool)` 이면 query
- `Initialize(ctx, ...scalar) error` 는 별도 special case

---

## 지원되는 상태 변수 타입

전역 persistent state는 scalar를 넘어 복합 타입까지 지원한다.

### 지원되는 상태 타입

- scalar
  - `string`
  - `bool`
  - `int`
  - `int64`
  - `uint64`
- named struct
- nested named struct
- `map[string]scalar`
- `map[string]named-struct`
- `[]scalar`
- `[]named-struct`
- struct field 안의 map
- struct field 안의 slice
- nested struct 안의 map/slice

즉 현재는 snapshot state에 struct/map/slice를 포함한 비교적 풍부한 상태를 저장할 수 있다.

---

## 지원되는 query 반환 타입

query result는 scalar보다 넓은 범위를 지원한다.

### 지원되는 query 반환 형태

- `T`
- `(T, bool)`

여기서 `T`는 아래 중 하나일 수 있다.

- scalar
- named struct
- `map[string]scalar`
- `map[string]named-struct`
- `[]scalar`
- `[]named-struct`

즉, 입력은 scalar-only지만 getter는 struct/map/slice를 그대로 반환할 수 있다.

---

## 입력 payload limit

현재 프로젝트는 scalar-only 입력 ABI를 유지하면서, `map[string]string` 입력 payload가 과도하게 커지는 것도 제한한다.

적용 대상:

- register `init_data`
- call operation `call_data`
- runtime direct execute request `CallData`
- runtime direct query request `CallData`
- digest query HTTP body에서 decode된 `callData`

현재 limit 값:

- 최대 entry 수: `64`
- key 최대 길이: `128 bytes`
- value 최대 길이: `16 KiB`
- 전체 key+value 총합: `64 KiB`

중요한 점:

- reserved key인 `function`도 일반 entry와 동일하게 계산된다
- `_sender`는 더 이상 query sender semantics로 해석되지 않으며, 들어오더라도 일반 unused key처럼 payload limit 계산에만 포함된다
- 길이는 문자 수가 아니라 **byte length** 기준이다
- composite input을 여는 기능이 아니라, 현재의 scalar string map 입력을 제한하는 정책이다

즉 아래처럼 key/value 수가 너무 많거나, key/value 하나가 너무 길거나, 전체 합이 너무 크면 contract 실행 전 단계에서 거부된다.

---

## 지원되지 않는 타입/형태

현재 명시적으로 지원하지 않는 것들은 아래와 같다.

### 타입 관련 비지원

- anonymous struct
- embedded field
- recursive struct
- mutually recursive struct
- non-string map key
- map value가 map인 형태
- map value가 slice인 형태
- slice element가 map인 형태
- slice element가 slice인 형태
- fixed array
- interface type
- pointer type
- func/channel 등 일반 Go 고급 타입

### ABI 관련 비지원

- write arg로 struct/map/slice 입력
- query arg로 struct/map/slice 입력
- query result 3개 이상 반환

### schema complexity 관련 비지원

typed Gno contract source가 문법적으로 맞더라도, schema가 너무 크거나 깊으면 admission 단계에서 거부된다.

현재 complexity limit:

- import 개수 최대: `16`
- 함수 개수 최대: `128`
- persistent global 개수 최대: `128`
- named struct 개수 최대: `64`
- struct field 개수 최대: `64` per struct
- 타입 nesting depth 최대: `16`
- 전체 schema node 수 최대: `4096`

이 제한은 "지원되는 타입이지만 너무 복잡한 contract"를 막기 위한 것이며, 기존 unsupported type 정책을 대체하는 것이 아니다.

---

## import 정책

현재 typed Gno contract는 **import allowlist** 정책을 사용한다.  
즉 일반 Go처럼 stdlib를 자유롭게 import할 수 없고, 아래 목록에 없는 import는 schema analysis 단계에서 거부된다.

현재 허용된 import는 정확히 아래뿐이다.

- `import "mitum/chain"`
- `strconv`
- `strings`
- `errors`
- `bytes`
- `encoding/hex`
- `encoding/base64`
- `unicode/utf8`

이외의 import는 모두 금지된다.

대표적인 금지 예시는 아래와 같다.

- `fmt`
- `regexp`
- `time`
- `math/rand`
- `net/url`
- `html`

특히 `math/rand`는 deterministic execution 정책 때문에 명시적으로 금지한다.  
같은 입력과 같은 chain state에서 실행 결과가 달라질 수 있는 요소는 contract language surface에서 열지 않는다.

`mitum/chain` 안에서 사용할 수 있는 대표 기능:

- `chain.WriteContext`
  - `ctx.GetSender()`
  - `ctx.GetContract()`
  - `ctx.GetHeight()`
  - `ctx.IsReadOnly()`
- `chain.QueryContext`
  - `ctx.GetContract()`
  - `ctx.GetHeight()`
  - `ctx.GetCurrentHeight()`
  - `ctx.IsReadOnly()`
- `chain.AccountExists(addr)`
- `chain.IsContractAccount(addr)`
- `chain.BalanceOf(addr, currency) (string, bool)`

`QueryContext`에는 `GetSender()`가 없다.

height는 아래처럼 구분한다.

- write의 `ctx.GetHeight()`: 현재 execution/block height
- query의 `ctx.GetHeight()`: 현재 query가 읽는 state/view height
- query의 `ctx.GetCurrentHeight()`: current chain head height

즉 write/register/call에서는 `ctx.GetHeight()`만 공식 height source이고, current head height는 query context method로만 제공된다. `chain.CurrentHeight()`는 더 이상 contract-facing ABI가 아니다.

balance 조회 ABI는 context method가 아니라 host native다.

```go
amount, ok := chain.BalanceOf(addr, currency)
```

의미:

- `amount`: decimal amount string
- `ok == true`: account, currency design, and balance state가 모두 존재함
- `ok == false`: account 없음, currency 없음, balance state 없음, 또는 address/currency 문자열이 유효하지 않음
- zero balance는 `"0", true`로 반환되어 not found와 구분된다

`BalanceOf`는 write/query 양쪽에서 호출 가능하다. 내부 state decode/type mismatch 같은 state corruption은 host native 실패로 처리되며, runtime panic surface sanitization 정책에 따라 raw 내부 detail은 client-visible error에 직접 노출되지 않는다. 현재 gas 값은 정밀 benchmark 전의 defensive baseline이다. `AccountExists`/`IsContractAccount`는 single state lookup tier `3000`, `BalanceOf`는 최대 세 state lookup tier `9000`으로 등록되어 있다.

즉, 현재 컨트랙트는 일반 Go 프로그램이라기보다 **제한된 typed Gno runtime 안에서 동작하는 코드**로 보는 것이 맞다.

---

## Query HTTP body limit

digest query HTTP endpoint는 decoded `callData` limit과 별도로 **raw HTTP body size limit**도 적용한다.

현재 query body 최대 크기:

- `128 KiB`

즉 query 요청은 아래 두 단계를 모두 통과해야 한다.

1. raw HTTP body size limit
2. decode 후 `map[string]string` payload limit

그래서 큰 JSON body는 `json.Unmarshal` 전에 먼저 거부될 수 있다.

---

## ReadOnly의 의미

`ctx.IsReadOnly()`는 "이 컨트랙트가 read-only contract로 배포되었다"는 뜻이 아니다.

이 값은 단지 **현재 실행이 query 경로인지 아닌지**를 알려주는 실행 컨텍스트 값이다.

- write/register/call 실행: `false`
- query 실행: `true`

즉, 컨트랙트의 영구 속성이 아니라 현재 실행 모드 정보다.

---

## 실용적인 작성 규칙

현재 프로젝트 기준으로 안전한 컨트랙트 작성 방식은 아래와 같다.

### 권장

- `package contract` 사용
- `Initialize(ctx) error` 정의
- 입력은 scalar 인자로만 받기
- 상태는 named struct / map / slice로 구성하기
- 복합 상태는 getter에서 struct/map/slice로 반환하기
- host ABI가 필요하면 `mitum/chain`만 사용하기

### 비권장

- stdlib를 일반 Go처럼 자유롭게 import하려고 시도하기
- anonymous struct를 상태/반환 타입으로 사용하기
- `init_data` / `call_data`에 매우 큰 string이나 과도한 수의 entry를 넣기
- 함수/struct/global을 불필요하게 많이 선언해 큰 schema를 만들기
- map 안에 map/slice를 직접 넣기
- slice 안에 map/slice를 직접 넣기
- 복합 타입을 write/query 입력으로 받으려 하기

---

## 요약

현재 정책은 다음 한 줄로 정리할 수 있다.

**입력은 scalar-only, 상태는 composite 가능, query 반환도 composite 가능하다.**

즉:

- write args: scalar only
- query args: scalar only
- input payload: size-limited string map
- state: struct/map/slice 지원
- query result: struct/map/slice 지원
- schema: complexity limit 적용
- host ABI: `mitum/chain` 중심으로 사용 가능

이 문서를 기준으로 컨트랙트를 작성하면 현재 runtime/ABI 정책과 가장 잘 맞는다.
