package common

type Env struct {
	data  map[MalTypeSymbol]MalType
	outer *Env
}

func NewEnv(outer *Env) Env {
	data := make(map[MalTypeSymbol]MalType)
	return Env{data, outer}
}

func NewEnvBind(outer *Env, binds []MalTypeSymbol, exprs []MalType) Env {
	env := NewEnv(outer)
	variadic := false
	for i := 0; i < len(binds); i++ {
		if variadic {
			env.Set(binds[i], MalTypeList(exprs[i-1:]))
			break
		}
		if binds[i] == MalTypeSymbol("&") {
			variadic = true
			continue
		}
		env.Set(binds[i], exprs[i])
	}
	return env
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
