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

func isMacroCall(ast common.MalType, env common.Env) (common.MalTypeTCOFunction, bool) {
	nilfun := common.MalTypeTCOFunction{}
	lst, ok := ast.(common.MalTypeList)
	if !ok {
		return nilfun, false
	}
	if len(lst) < 1 {
		return nilfun, false
	}
	first := lst[0]
	sym, ok := first.(common.MalTypeSymbol)
	if !ok {
		return nilfun, false
	}
	ref, ok := env.Get(sym)
	if !ok {
		return nilfun, false
	}
	fun, ok := ref.(common.MalTypeTCOFunction)
	if !ok {
		return fun, false
	}
	return fun, fun.IsMacro
}

func macroexpand(ast common.MalType, env common.Env) (common.MalType, error) {
	for {
		fun, ok := isMacroCall(ast, env)
		if !ok {
			break
		}
		ast_as_list := ast.(common.MalTypeList)
		var err error
		ast, err = fun.Fn(ast_as_list[1:])
		if err != nil {
			return nil, err
		}
	}
	return ast, nil
}

func apply(lst common.MalTypeList, env common.Env) (common.MalType, common.Env, bool, error) {
	evaluated, err := eval_ast(lst, env)
	if err != nil {
		return nil, env, false, err
	}
	evaluated_list := evaluated.(common.MalTypeList)
	fun, ok := evaluated_list[0].(common.MalTypeFunction)
	if ok {
		ret, err := fun(evaluated_list[1:])
		return ret, env, false, err
	}
	
	tco_fun, ok := evaluated_list[0].(common.MalTypeTCOFunction)
	if ok {
		new_env, err := common.NewEnvBind(&tco_fun.Env, tco_fun.Params, evaluated_list[1:])
		if err != nil {
			return nil, env, false, err
		}
		return tco_fun.Ast, new_env, true, nil
	}
	
	return nil, env, false, errors.New(fmt.Sprintf("cannot call object of type %T", evaluated_list[0]))
}

func apply_def(lst common.MalTypeList, isMacro bool, env common.Env) (common.MalType, error) {
	if len(lst) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid number of parameters for def! call, expected 2 but got %d", len(lst)))
	}
	first, ok := lst[0].(common.MalTypeSymbol)
	if !ok {
		return nil, errors.New(fmt.Sprintf("expected symbol in call to def! or defmacro! but got %T", lst[0]))
	}
	second, err := EVAL(lst[1], env)
	if err != nil {
		return nil, err
	}
	if isMacro {
		second_as_fun, ok := second.(common.MalTypeTCOFunction)
		if !ok {
			return nil, errors.New(fmt.Sprintf("expected second parameter of defmacro! to be a function but got %T", second))
		}
		second_as_fun.IsMacro = true
		env.Set(first, second_as_fun)
		return second_as_fun, nil
	} else {
		env.Set(first, second)
		return second, nil
	}
}

func apply_let(lst common.MalTypeList, env common.Env) (common.MalType, common.Env, error) {
	if len(lst) != 2 {
		return nil, env, errors.New(fmt.Sprintf("invalid number of parameters for let* call, expected 2 but got %d", len(lst)))
	}
	new_env := common.NewEnv(&env)

	var bindings []common.MalType
	switch b := lst[0].(type) {
	case common.MalTypeList:
		bindings = []common.MalType(b)
	case common.MalTypeVector:
		bindings = []common.MalType(b)
	default:
		return nil, env, errors.New(fmt.Sprintf("expected sequence in call to let* but got %T", lst[0]))
	}

	if len(bindings) % 2 != 0 {
		return nil, env, errors.New(fmt.Sprintf("expected bindings list to have an even number of elements, has %d", len(bindings)))
	}
	for i := 0; i < len(bindings); i += 2 {
		second, err := EVAL(bindings[i+1], new_env)
		if err != nil {
			return nil, env, err
		}
		first, ok := bindings[i].(common.MalTypeSymbol)
		if !ok {
			return nil, env, errors.New(fmt.Sprintf("expected odd binding elements to be %T but found a %T", first, bindings[i]))
		}
		new_env.Set(first, second)
	}
	return lst[1], new_env, nil
}

func apply_do(lst common.MalTypeList, env common.Env) (common.MalType, error) {
	for _, element := range lst[:len(lst)-1] {
		_, err := EVAL(element, env)
		if err != nil {
			return nil, err
		}
	}
	return lst[len(lst)-1], nil
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
		return lst[2], nil
	} else {
		return lst[1], nil
	}
}

func apply_fn(lst common.MalTypeList, env common.Env) (common.MalTypeTCOFunction, error) {
	tco_fn := common.MalTypeTCOFunction{
		Ast: nil,
		Params: nil,
		Env: env,
		IsMacro: false,
		Fn: nil,
	}
	
	if len(lst) != 2 {
		return tco_fn, errors.New(fmt.Sprintf("invalid number of parameters for fn* call, expected 2 but got %d", len(lst)))
	}
	tco_fn.Ast = lst[1]
	
	var names_array []common.MalType
	names_list, ok := lst[0].(common.MalTypeList)
	if ok {
		names_array = []common.MalType(names_list)
	} else {
		names_vector, ok := lst[0].(common.MalTypeVector)
		if !ok {
			return tco_fn, errors.New(fmt.Sprintf("expected first parameter of fn* to be of type %T or %T but was %T", names_list, names_vector, lst[0]))
		}
		names_array = []common.MalType(names_vector)
	}
	
	names := make([]common.MalTypeSymbol, len(names_array))
	for i, name := range names_array {
		n, ok := name.(common.MalTypeSymbol)
		if !ok {
			return tco_fn, errors.New(fmt.Sprintf("expected first parameter of fn* to be a list of %T but found a %T", n, name))
		}
		names[i] = n
	}
	tco_fn.Params = names

	tco_fn.Fn = common.MalTypeFunction(func(args []common.MalType) (common.MalType, error) {
		new_env, err := common.NewEnvBind(&env, names, args)
		if err != nil {
			return nil, err
		}
		return EVAL(lst[1], new_env)
	})
	
	return tco_fn, nil
}

func apply_quote(lst common.MalTypeList) (common.MalType, error) {
	if len(lst) != 1 {
		return nil, errors.New("expected one argument to quote")
	}
	return lst[0], nil
}

func apply_quasiquote(args common.MalTypeList) (common.MalType, error) {
	if len(args) != 1 {
		return nil, errors.New("expected one argument to quote")
	}
	ast := args[0]
	is_vec := false
	var ast_as_arr []common.MalType
	var ok bool
	ast_as_arr, ok = ast.(common.MalTypeList)
	if !ok {
		ast_as_arr, ok = ast.(common.MalTypeVector)
		is_vec = ok
	}
	if ok {
		if !is_vec && len(ast_as_arr) > 0 {
			if ast_as_arr[0] == common.MalTypeSymbol("unquote") {
				if len(ast_as_arr) == 2 {
					return ast_as_arr[1], nil
				} else {
					return nil, errors.New("too many arguments for unquote")
				}
			}
		}
		res := make([]common.MalType, 0)
		for i := len(ast_as_arr)-1; i >= 0; i-- {
			elt := ast_as_arr[i]
			elt_as_lst, ok := elt.(common.MalTypeList)
			if ok {
				if len(elt_as_lst) > 0 {
					if elt_as_lst[0] == common.MalTypeSymbol("splice-unquote") {
						if len(elt_as_lst) == 2 {
							new_res := make([]common.MalType, 3)
							new_res[0] = common.MalTypeSymbol("concat")
							new_res[1] = elt_as_lst[1]
							new_res[2] = common.MalTypeList(res)
							res = new_res
							continue
						} else {
							return nil, errors.New("too many arguments for splice-unquote")
						}
					}
				}
			}
			args := []common.MalType{elt}
			rest, err := apply_quasiquote(args)
			if err != nil {
				return nil, err
			}
			new_res := make([]common.MalType, 3)
			new_res[0] = common.MalTypeSymbol("cons")
			new_res[1] = rest
			new_res[2] = common.MalTypeList(res)
			res = new_res
		}
		if is_vec {
			new_res := make([]common.MalType, 2)
			new_res[0] = common.MalTypeSymbol("vec")
			new_res[1] = common.MalTypeList(res)
			res = new_res
		}
		return common.MalTypeList(res), nil
	}
	_, ok = ast.(common.MalTypeHashMap)
	if !ok {
		_, ok = ast.(common.MalTypeSymbol)
	}
	if ok {
		res := make([]common.MalType, 2)
		res[0] = common.MalTypeSymbol("quote")
		res[1] = ast
		return common.MalTypeList(res), nil
	}
	return ast, nil
}

func apply_macroexpand(ast common.MalTypeList, env common.Env) (common.MalType, error) {
	if len(ast) > 1 {
		return nil, errors.New("expected one argument to macroexpand")
	}
	return macroexpand(ast[0], env)
}

func apply_dir(ast common.MalTypeList, env common.Env) (common.MalType, error) {
	if len(ast) > 0 {
		return nil, errors.New("expected no argument to dir")
	}
	for k,v := range env.GetData() {
		fmt.Printf("%s: %s\n", string(k), common.PrStr(v, true))
	}
	return common.MalTypeNil{}, nil
}

func EVAL(ast common.MalType, env common.Env) (common.MalType, error) {
	for {
		var err error
		ast, err = macroexpand(ast, env)
		if err != nil {
			return nil, err
		}
		
		switch l := ast.(type) {
		case common.MalTypeList:
			if len(l) == 0 {
				return ast, nil
			} else {
				switch l[0] {
				case common.MalTypeSymbol("def!"):
					return apply_def(l[1:], false, env)
				case common.MalTypeSymbol("defmacro!"):
					return apply_def(l[1:], true, env)
				case common.MalTypeSymbol("let*"):
					ast, env, err = apply_let(l[1:], env)
					if err != nil {
						return nil, err
					}
					continue
				case common.MalTypeSymbol("do"):
					ast, err = apply_do(l[1:], env)
					if err != nil {
						return nil, err
					}
					continue
				case common.MalTypeSymbol("if"):
					ast, err = apply_if(l[1:], env)
					if err != nil {
						return nil, err
					}
					continue
				case common.MalTypeSymbol("fn*"):
					return apply_fn(l[1:], env)
				case common.MalTypeSymbol("quote"):
					return apply_quote(l[1:])
				case common.MalTypeSymbol("quasiquote"):
					ast, err = apply_quasiquote(l[1:])
					if err != nil {
						return nil, err
					}
					continue
				case common.MalTypeSymbol("quasiquoteexpand"):
					return apply_quasiquote(l[1:])
				case common.MalTypeSymbol("macroexpand"):
					return apply_macroexpand(l[1:], env)
				case common.MalTypeSymbol("dir"):
					return apply_dir(l[1:], env)
				default:
					var loop bool
					ast, env, loop, err = apply(l, env)
					if err != nil {
						return nil, err
					}
					if loop {
						continue
					}
					return ast, nil
				}
			}
		default:
			return eval_ast(ast, env)
		}
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
	repl_env := common.NewEnv(nil)
	for key, value := range common.Ns() {
		repl_env.Set(common.MalTypeSymbol(key), common.MalTypeFunction(value))
	}
	repl_env.Set(common.MalTypeSymbol("eval"), common.MalTypeFunction(func(args []common.MalType) (common.MalType, error) {
		if len(args) != 1 {
			return nil, errors.New("expected exactly 1 argument")
		}
		return EVAL(args[0], repl_env)
	}))

	rep("(def! not (fn* (a) (if a false true)))", repl_env)
	rep("(def! load-file (fn* (f) (eval (read-string (str \"(do \" (slurp f) \"\nnil)\")))))", repl_env)
	rep("(defmacro! cond (fn* (& xs) (if (> (count xs) 0) (list 'if (first xs) (if (> (count xs) 1) (nth xs 1) (throw \"odd number of forms to cond\")) (cons 'cond (rest (rest xs)))))))", repl_env)

	argv := make([]common.MalType, max(len(os.Args)-2, 0))
	if len(os.Args) > 2 {
		for i, arg := range os.Args[2:] {
			argv[i] = common.MalTypeString(arg)
		}
	}
	repl_env.Set(common.MalTypeSymbol("*ARGV*"), common.MalTypeList(argv))

	if len(os.Args) > 1 {
		_, err := rep(fmt.Sprintf("(load-file \"%s\")", os.Args[1]), repl_env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	rl, err := readline.New("user> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()
					
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		
		ret, err := rep(line, repl_env)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Println(ret)
		}
	}
}
