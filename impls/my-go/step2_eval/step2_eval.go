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

func eval_ast_list(elements []common.MalType, env map[common.MalTypeSymbol]common.MalTypeFunction) ([]common.MalType, error) {
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

func eval_ast(ast common.MalType, env map[common.MalTypeSymbol]common.MalTypeFunction) (common.MalType, error) {
	switch a := ast.(type) {
	case common.MalTypeSymbol:
		v, ok := env[a]
		if !ok {
			return nil, errors.New(fmt.Sprintf("symbol '%s' undefined", a))
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

func EVAL(ast common.MalType, env map[common.MalTypeSymbol]common.MalTypeFunction) (common.MalType, error) {
	switch l := ast.(type) {
	case common.MalTypeList:
		if len(l) == 0 {
			return ast, nil
		} else {
			evaluated, err := eval_ast(ast, env)
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
	default:
		return eval_ast(ast, env)
	}
		
}

func PRINT(input common.MalType) string {
	return common.PrStr(input, true)
}

func rep(input string, env map[common.MalTypeSymbol]common.MalTypeFunction) (string, error) {
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

	repl_env := map[common.MalTypeSymbol]common.MalTypeFunction {
		common.MalTypeSymbol("+"): sum_fun,
		common.MalTypeSymbol("-"): sub_fun,
		common.MalTypeSymbol("*"): mul_fun,
		common.MalTypeSymbol("/"): div_fun,
	}

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
