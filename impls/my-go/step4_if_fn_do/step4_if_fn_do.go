package main

import (
	"errors"
	"fmt"
	"os"
	"example.com/mal/common"
	"github.com/chzyer/readline"
)

func READ(input string) (common.MalType, error) {
	return common.ReadStr(input)
}

func eval_ast_list(elements []common.MalType, env common.Env) ([]common.MalType, error) {
	new_list := make([]common.MalType, len(elements))
	for i, element := range elements {
		evaled_element, err := EVAL(element, env)
		if err != nil {
			return nil, err
		}
		new_list[i] = evaled_element
	}
	return new_list, nil
}

func eval_ast(ast common.MalType, env common.Env) (common.MalType, error) {
	switch a := ast.(type) {
	case common.MalTypeSymbol:
		v, ok := env.Get(a)
		if !ok {
			return nil, errors.New(fmt.Sprintf("symbol '%s' not found", a))
		}
		return v, nil
	case common.MalTypeList:
		new_list, err := eval_ast_list(a, env)
		if err != nil {
			return nil, err
		}
		return common.MalTypeList(new_list), nil
	case common.MalTypeVector:
		new_list, err := eval_ast_list(a, env)
		if err != nil {
			return nil, err
		}
		return common.MalTypeVector(new_list), nil
	case common.MalTypeHashMap:
		var keys []common.MalType
		var values []common.MalType
		for key, value := range a {
			keys = append(keys, key)
			values = append(values, value)
		}
		new_keys, err := eval_ast_list(keys, env)
		if err != nil {
			return nil, err
		}
		new_values, err := eval_ast_list(values, env)
		if err != nil {
			return nil, err
		}
		new_map := make(common.MalTypeHashMap)
		for i, key := range new_keys {
			value := new_values[i]
			new_map[key] = value
		}
		return common.MalTypeHashMap(new_map), nil
	default:
		return ast, nil
	}
}

func apply(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	evaluated, err := eval_ast(lst, env)
	if err != nil {
		return nil, err
	}
	evaluated_list := evaluated.(common.MalTypeList)
	fun, ok := evaluated_list[0].(common.MalTypeFunction)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot call object of type %T", evaluated_list[0]))
	}
	return fun(evaluated_list[1:])
}

func apply_def(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	if len(lst) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for def! call, expected 2 but got %d", len(lst)))
	}
	first, ok := lst[0].(common.MalTypeSymbol)
	if !ok {
		return nil, errors.New(fmt.Sprintf("expected symbol in call to def! but got %T", lst[0]))
	}
	second, err := EVAL(lst[1], env)
	if err != nil {
		return nil, err
	}
	env.Set(first, second)
	return second, nil
}

func apply_let(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	if len(lst) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for let* call, expected 2 but got %d", len(lst)))
	}
	new_env := common.NewEnv(&env)

	var bindings []common.MalType
	switch b := lst[0].(type) {
	case common.MalTypeList:
		bindings = []common.MalType(b)
	case common.MalTypeVector:
		bindings = []common.MalType(b)
	default:
		return nil, errors.New(fmt.Sprintf("expected sequence in call to let* but got %T", lst[0]))
	}

	if len(bindings) % 2 != 0 {
		return nil, errors.New(fmt.Sprintf("expected bindings list to have an even number of elements, has %d", len(bindings)))
	}
	for i := 0; i < len(bindings); i += 2 {
		second, err := EVAL(bindings[i+1], new_env)
		if err != nil {
			return nil, err
		}
		first, ok := bindings[i].(common.MalTypeSymbol)
		if !ok {
			return nil, errors.New(fmt.Sprintf("expected odd binding elements to be %T but found a %T", first, bindings[i]))
		}
		new_env.Set(first, second)
	}
	return EVAL(lst[1], new_env)
}

func apply_do(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	var last common.MalType = common.MalTypeNil{}
	for _, element := range lst {
		res, err := EVAL(element, env)
		if err != nil {
			return nil, err
		}
		last = res
	}
	return last, nil
}

func apply_if(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	if len(lst) < 2 || len(lst) > 3 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for if call, expected 2 or 3 but got %d", len(lst)))
	}
	condition, err := EVAL(lst[0], env)
	if err != nil {
		return nil, err
	}
	
	nilinst := common.MalTypeNil{}
	if condition == common.MalTypeBoolean(false) || condition == nilinst {
		if len(lst) == 2 {
			return nilinst, nil
		}
		return EVAL(lst[2], env)
	} else {
		return EVAL(lst[1], env)
	}
}

func apply_fn(lst common.MalTypeList, env common.Env) (common.MalTypeFunction, error) {
	if len(lst) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for fn* call, expected 2 but got %d", len(lst)))
	}
	var names_array []common.MalType
	names_list, ok := lst[0].(common.MalTypeList)
	if ok {
		names_array = []common.MalType(names_list)
	} else {
		names_vector, ok := lst[0].(common.MalTypeVector)
		if !ok {
			return nil, errors.New(fmt.Sprintf("expected first parameter of fn* to be of type %T or %T but was %T", names_list, names_vector, lst[0]))
		}
		names_array = []common.MalType(names_vector)
	}
	names := make([]common.MalTypeSymbol, len(names_array))
	for i, name := range names_array {
		n, ok := name.(common.MalTypeSymbol)
		if !ok {
			return nil, errors.New(fmt.Sprintf("expected first parameter of fn* to be a list of %T but found a %T", n, name))
		}
		names[i] = n
	}
	return common.MalTypeFunction(func(args []common.MalType) (common.MalType, error) {
		new_env, err := common.NewEnvBind(&env, names, args)
		if err != nil {
			return nil, err
		}
		return EVAL(lst[1], new_env)
	}), nil
}

func EVAL(ast common.MalType, env common.Env) (common.MalType, error) {
	switch l := ast.(type) {
	case common.MalTypeList:
		if len(l) == 0 {
			return ast, nil
		} else {
			switch l[0] {
			case common.MalTypeSymbol("def!"):
				return apply_def(l[1:], env)
			case common.MalTypeSymbol("let*"):
				return apply_let(l[1:], env)
			case common.MalTypeSymbol("do"):
				return apply_do(l[1:], env)
			case common.MalTypeSymbol("if"):
				return apply_if(l[1:], env)
			case common.MalTypeSymbol("fn*"):
				return apply_fn(l[1:], env)
			default:
				return apply(l, env)
			}
		}
	default:
		return eval_ast(ast, env)
	}
		
}

func PRINT(input common.MalType) string {
	return common.PrStr(input, true)
}

func rep(input string, env common.Env) (string, error) {
	read_val, err := READ(input)
	if err != nil {
		return "", err
	}
	eval_val, err := EVAL(read_val, env)
	if err != nil {
		return "", err
	}
	return PRINT(eval_val), nil
}

func main() {
	rl, err := readline.New("user> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	repl_env := common.NewEnv(nil)
	for key, value := range common.Ns() {
		repl_env.Set(common.MalTypeSymbol(key), common.MalTypeFunction(value))
	}

	rep("(def! not (fn* (a) (if a false true)))", repl_env)
					
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		
		ret, err := rep(line, repl_env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err);
		} else {
			fmt.Println(ret)
		}
	}
}
