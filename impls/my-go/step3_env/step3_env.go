package main;

import (
	"os"
	"fmt"
	"errors"
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
		new_list, err := eval_ast_list(a.List, env)
		if err != nil {
			return nil, err
		}
		return common.NewMalList(new_list), nil
	case common.MalTypeVector:
		new_list, err := eval_ast_list(a.Vector, env)
		if err != nil {
			return nil, err
		}
		return common.NewMalVector(new_list), nil
	case common.MalTypeHashMap:
		var keys []common.MalType
		var values []common.MalType
		for key, value := range a.HashMap {
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
		new_map := make(map[common.MalType]common.MalType)
		for i, key := range new_keys {
			value := new_values[i]
			new_map[key] = value
		}
		return common.NewMalHashMap(new_map), nil
	default:
		return ast, nil
	}
}

func apply(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	evaluated, err := eval_ast(lst, env)
	if err != nil {
		return nil, err
	}
	evaluated_list := evaluated.(common.MalTypeList).List
	fun, ok := evaluated_list[0].(common.MalTypeFunction)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot call object of type %T", evaluated_list[0]))
	}
	return fun.Func(evaluated_list[1:])
}

func apply_def(lst []common.MalType, env common.Env) (common.MalType, error) {
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

func apply_let(lst []common.MalType, env common.Env) (common.MalType, error) {
	if len(lst) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for let* call, expected 2 but got %d", len(lst)))
	}
	new_env := common.NewEnv(&env)

	var bindings []common.MalType
	switch b := lst[0].(type) {
	case common.MalTypeList:
		bindings = b.List
	case common.MalTypeVector:
		bindings = b.Vector
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

func EVAL(ast common.MalType, env common.Env) (common.MalType, error) {
	switch l := ast.(type) {
	case common.MalTypeList:
		if len(l.List) == 0 {
			return ast, nil
		} else {
			switch l.List[0] {
			case common.MalTypeSymbol("def!"):
				return apply_def(l.List[1:], env)
			case common.MalTypeSymbol("let*"):
				return apply_let(l.List[1:], env)
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

func assert_int_args(args []common.MalType) error {
	for i, element := range args {
		res, ok := element.(common.MalTypeInteger)
		if !ok {
			return errors.New(fmt.Sprintf("argument %d expected to be of type %T but was %T", i, res, element))
		}
	}
	return nil
}

func sum_fun(args []common.MalType) (common.MalType, error) {
	err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(0)
	for _, element := range args {
		int_element := element.(common.MalTypeInteger)
		res += int64(int_element)
	}

	return common.MalTypeInteger(res), nil
}

func sub_fun(args []common.MalType) (common.MalType, error) {
	if len(args) == 0 {
		return common.MalTypeInteger(0), nil
	}
	err := assert_int_args(args)
	
	sum_res, err := sum_fun(args[1:])
	if err != nil {
		return nil, err
	}

	return common.MalTypeInteger(args[0].(common.MalTypeInteger) - sum_res.(common.MalTypeInteger)), nil
}

func mul_fun(args []common.MalType) (common.MalType, error) {
	err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(1)
	for _, element := range args {
		int_element := element.(common.MalTypeInteger)
		res *= int64(int_element)
	}

	return common.MalTypeInteger(res), nil
}

func div_fun(args []common.MalType) (common.MalType, error) {
	if len(args) == 0 {
		return nil, errors.New("invalid number of arguments")
	}
	
	err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	if len(args) == 1 {
		return common.MalTypeInteger(1 / int64(args[0].(common.MalTypeInteger))), nil
	}

	rest, err := mul_fun(args[1:])
	if err != nil {
		return nil, err
	}
	return common.MalTypeInteger(args[0].(common.MalTypeInteger) / rest.(common.MalTypeInteger)), nil
}

func main() {
	rl, err := readline.New("user> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	repl_env := common.NewEnv(nil)
	repl_env.Set(common.MalTypeSymbol("+"), common.NewMalFunction(sum_fun))
	repl_env.Set(common.MalTypeSymbol("-"), common.NewMalFunction(sub_fun))
	repl_env.Set(common.MalTypeSymbol("*"), common.NewMalFunction(mul_fun))
	repl_env.Set(common.MalTypeSymbol("/"), common.NewMalFunction(div_fun))
					
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
