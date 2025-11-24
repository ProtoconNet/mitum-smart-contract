package util

type S interface {
	Name() string
}

type S1 struct {
	N string
}

func (s S1) Name() string {
	return s.N
}

type APICollection func(int, string) (interface{}, error)
type GetAccountStateFunc func(string) (bool, error)
type GetDataStateFunc func(string) (map[string]interface{}, error)
type GetSenderFunc func() string
type GetCallDataFunc func() map[string]string

type ContractContext struct {
	GetAccountState GetAccountStateFunc
	GetDataState    GetDataStateFunc
	GetSender       GetSenderFunc
	GetCallData     GetCallDataFunc
}
