package common

type Env struct {
	data  map[MalTypeSymbol]MalType
	outer *Env
}

func NewEnv(outer *Env) Env {
	data := make(map[MalTypeSymbol]MalType)
	return Env{data, outer}
}

func (e *Env) Set(key MalTypeSymbol, value MalType) {
	e.data[key] = value
}

func (e Env) Get(key MalTypeSymbol) (MalType, bool) {
	curr := e
	for {
		ret, ok := curr.data[key]
		if ok {
			return ret, true
		} else if curr.outer == nil {
			return nil, false
		}
		curr = *curr.outer
	}
}
