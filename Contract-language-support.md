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
- `Initialize(ctx chain.ContractContext, ...scalar) error`

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
  - `func X(ctx chain.ContractContext, a string, b int64) error`
- query 함수
  - `func X(ctx chain.ContractContext, name string) bool`
  - `func X(ctx chain.ContractContext, index int) (string, bool)`

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
func X(ctx chain.ContractContext, ...scalar) error
```

조건:

- exported 함수
- 첫 번째 인자가 `chain.ContractContext`
- 함수 이름이 `Initialize`가 아님
- 반환값이 **오직 `error` 1개**

즉 반환형이 `error` 하나이면 write 함수로 분류된다.

### query 함수

query 함수는 아래 형태를 따른다.

```go
func X(ctx chain.ContractContext, ...scalar) T
func X(ctx chain.ContractContext, ...scalar) (T, bool)
```

조건:

- exported 함수
- 첫 번째 인자가 `chain.ContractContext`
- 함수 이름이 `Initialize`가 아님
- 반환값이
  - 1개이거나
  - 2개인데 두 번째가 `bool`

즉 반환형이 `T` 또는 `(T, bool)`이면 query 함수로 분류된다.

### Initialize는 예외

`Initialize`는 write/query 일반 규칙과 별도로 취급되는 특수 함수다.

형태:

```go
func Initialize(ctx chain.ContractContext, ...scalar) error
```

이 함수는 컨트랙트 초기화 전용 entrypoint이며, 일반 write 함수와는 별개로 처리된다.

중요:

- `Initialize`는 scalar arg를 받을 수 있다
- 전달은 순서 기반이 아니라 **이름 기반**이다
- register payload의 `init_data` key가 `Initialize` 파라미터 이름과 일치해야 한다

예:

```go
func Initialize(ctx chain.ContractContext, owner string, label string, limit int64) error
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

### 요약

- `... -> error` 이면 write
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

- `chain.ContractContext`
- `ctx.GetSender()`
- `ctx.GetContract()`
- `ctx.GetHeight()`
- `ctx.IsReadOnly()`
- `chain.AccountExists(addr)`
- `chain.IsContractAccount(addr)`

즉, 현재 컨트랙트는 일반 Go 프로그램이라기보다 **제한된 typed Gno runtime 안에서 동작하는 코드**로 보는 것이 맞다.

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
- state: struct/map/slice 지원
- query result: struct/map/slice 지원
- host ABI: `mitum/chain` 중심으로 사용 가능

이 문서를 기준으로 컨트랙트를 작성하면 현재 runtime/ABI 정책과 가장 잘 맞는다.
