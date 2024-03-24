package common

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func make_type_err(found MalType, expected MalType) error {
	return errors.New(fmt.Sprintf("expected argument of type %T but found %T", expected, found))
}

func assert_int_arg(arg MalType) (MalTypeInteger, error) {
	res, ok := arg.(MalTypeInteger)
	if ok {
		return res, nil
	}
	return res, make_type_err(arg, res)
}

func assert_string_arg(arg MalType) (MalTypeString, error) {
	res, ok := arg.(MalTypeString)
	if ok {
		return res, nil
	}
	return res, make_type_err(arg, res)
}

func assert_list_arg(arg MalType) (MalTypeList, error) {
	res, ok := arg.(MalTypeList)
	if ok {
		return res, nil
	}
	return res, make_type_err(arg, res)
}

func assert_list_or_vec_arg(arg MalType) ([]MalType, error) {
	res, ok := arg.(MalTypeList)
	if ok {
		return res, nil
	}
	res_vec, ok := arg.(MalTypeVector)
	if ok {
		return res_vec, nil
	}
	return res, make_type_err(arg, res)
}

func assert_atom_arg(arg MalType) (MalTypeAtom, error) {
	res, ok := arg.(MalTypeAtom)
	if ok {
		return res, nil
	}
	return res, make_type_err(arg, res)
}

func assert_function_arg(arg MalType) (MalTypeFunction, error) {
	res, ok := arg.(MalTypeFunction)
	if ok {
		return res, nil
	}
	tco, ok := arg.(MalTypeTCOFunction)
	if ok {
		return tco.Fn, nil
	}
	return res, make_type_err(arg, res)
}

func assert_int_args(args []MalType) ([]MalTypeInteger, error) {
	res := make([]MalTypeInteger, len(args))
	for i, arg := range args {
		n, err := assert_int_arg(arg)
		if err != nil {
			return nil, err
		}
		res[i] = n
	}
	return res, nil
}

func assert_list_or_vec_args(args []MalType) ([][]MalType, error) {
	res := make([][]MalType, len(args))
	for i, arg := range args {
		l, err := assert_list_or_vec_arg(arg)
		if err != nil {
			return nil, err
		}
		res[i] = l
	}
	return res, nil
}

func assert_len_args(args []MalType, min int, max int) error {
	if (min > 0 && len(args) < min) || (max > 0 && len(args) > max) {
		var err string
		if min > 0 && max > 0 {
			if min == max {
				err = fmt.Sprintf("expected exactly %d arguments but found %d", min, len(args))
			} else {
				err = fmt.Sprintf("expected between %d and %d arguments but found %d", min, max, len(args))
			}
		} else if min > 0 {
			err = fmt.Sprintf("expected at least %d arguments but found %d", min, len(args))
		} else if max > 0 {
			err = fmt.Sprintf("expected less than %d arguments but found %d", max, len(args))
		} else {
			return nil
		}
		return errors.New(err)
	}
	return nil
}

func assert_nonempty_args(args []MalType) error {
	return assert_len_args(args, 1, -1)
}

func sum_fun(args []MalType) (MalType, error) {
	ints, err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(0)
	for _, n := range ints {
		res += int64(n)
	}

	return MalTypeInteger(res), nil
}

func sub_fun(args []MalType) (MalType, error) {
	if len(args) == 0 {
		return MalTypeInteger(0), nil
	}

	first, err := assert_int_arg(args[0])
	if err != nil {
		return nil, err
	}
	
	sum_res, err := sum_fun(args[1:])
	if err != nil {
		return nil, err
	}

	return MalTypeInteger(first - sum_res.(MalTypeInteger)), nil
}

func mul_fun(args []MalType) (MalType, error) {
	ints, err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(1)
	for _, n := range ints {
		res *= int64(n)
	}

	return MalTypeInteger(res), nil
}

func div_fun(args []MalType) (MalType, error) {
	err := assert_nonempty_args(args)
	if err != nil {
		return nil, err
	}
	ints, err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	if len(args) == 1 {
		return MalTypeInteger(1 / ints[0]), nil
	}

	rest, err := mul_fun(args[1:])
	if err != nil {
		return nil, err
	}
	return MalTypeInteger(ints[0] / rest.(MalTypeInteger)), nil
}

func pr_str_fun(args []MalType) (MalType, error) {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = PrStr(arg, true)
	}
	return MalTypeString(strings.Join(strs, " ")), nil
}

func str_fun(args []MalType) (MalType, error) {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = PrStr(arg, false)
	}
	return MalTypeString(strings.Join(strs, "")), nil
}

func prn_fun(args []MalType) (MalType, error) {
	str, err := pr_str_fun(args)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(str.(MalTypeString)))
	return MalTypeNil{}, nil
}

func println_fun(args []MalType) (MalType, error) {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = PrStr(arg, false)
	}
	fmt.Println(strings.Join(strs, " "))
	return MalTypeNil{}, nil
}

func list_fun(args []MalType) (MalType, error) {
	return MalTypeList(args), nil
}

func list_pred_fun(args []MalType) (MalType, error) {
	err := assert_nonempty_args(args)
	if err != nil {
		return nil, err
	}
	switch args[0].(type) {
	case MalTypeList:
		return MalTypeBoolean(true), nil
	default:
		return MalTypeBoolean(false), nil
	}
}

func count_fun(args []MalType) (MalType, error) {
	err := assert_nonempty_args(args)
	if err != nil {
		return nil, err
	}
	switch l := args[0].(type) {
	case MalTypeList:
		return MalTypeInteger(len(l)), nil
	case MalTypeVector:
		return MalTypeInteger(len(l)), nil
	case MalTypeNil:
		return MalTypeInteger(0), nil
	default:
		return nil, errors.New(fmt.Sprintf("expected argument to be a sequence but is %T", args[0]))
	}
}

func empty_pred_fun(args []MalType) (MalType, error) {
	count, err := count_fun(args)
	if err != nil {
		return nil, err
	}
	return MalTypeBoolean(int64(count.(MalTypeInteger)) == 0), nil
}

func get_as_mal_array(x MalType) ([]MalType, bool) {
	switch y := x.(type) {
	case MalTypeList:
		return []MalType(MalTypeList(y)), true
	case MalTypeVector:
		return []MalType(MalTypeVector(y)), true
	default:
		return nil, false
	}
}

func eq_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 2, -1)
	if err != nil {
		return nil, err
	}

	new_args := make([]MalType, 2)
	
	first := args[0]
	second := args[1]

	first_arr, ok := get_as_mal_array(first)
	if ok {
		second_arr, ok := get_as_mal_array(second)
		if !ok {
			return MalTypeBoolean(false), nil
		}
		if len(first_arr) != len(second_arr) {
			return MalTypeBoolean(false), nil
		}
		for i := 0; i < len(first_arr); i++ {
			new_args[0] = first_arr[i]
			new_args[1] = second_arr[i]
			res, err := eq_fun(new_args)
			if err != nil {
				return nil, err
			}
			if res == MalTypeBoolean(false) {
				return res, nil
			}
		}
		return MalTypeBoolean(true), nil
	}

	first_map, ok := first.(MalTypeHashMap)
	if ok {
		second_map, ok := second.(MalTypeHashMap)
		if !ok {
			return MalTypeBoolean(false), nil
		}
		if len(first_map) != len(second_map) {
			return MalTypeBoolean(false), nil
		}
		for key, value1 := range first_map {
			value2, ok := second_map[key]
			if !ok {
				return MalTypeBoolean(false), nil
			}
			new_args[0] = value1
			new_args[1] = value2
			res, err := eq_fun(new_args)
			if err != nil {
				return nil, err
			}
			if res == MalTypeBoolean(false) {
				return res, nil
			}
		}
		return MalTypeBoolean(true), nil
	}

	switch f := first.(type) {
	case MalTypeInteger:
		s, ok := second.(MalTypeInteger)
		if ok && int64(f) == int64(s) {
			return MalTypeBoolean(true), nil
		}
	case MalTypeNil:
		_, ok := second.(MalTypeNil)
		if ok {
			return MalTypeBoolean(true), nil
		}
	case MalTypeBoolean:
		s, ok := second.(MalTypeBoolean)
		if ok && bool(f) == bool(s) {
			return MalTypeBoolean(true), nil
		}
	case MalTypeKeyword:
		s, ok := second.(MalTypeKeyword)
		if ok && string(f) == string(s) {
			return MalTypeBoolean(true), nil
		}
	case MalTypeString:
		s, ok := second.(MalTypeString)
		if ok && string(f) == string(s) {
			return MalTypeBoolean(true), nil
		}
	case MalTypeSymbol:
		s, ok := second.(MalTypeSymbol)
		if ok && string(f) == string(s) {
			return MalTypeBoolean(true), nil
		}
	case MalTypeFunction:
		return MalTypeBoolean(false), nil
	}

	return MalTypeBoolean(false), nil
}

func get_two_ints_for_cmp(args []MalType) (int64, int64, error) {
	err := assert_len_args(args, 2, -1)
	if err != nil {
		return 0, 0, err
	}
	ints, err := assert_int_args(args)
	if err != nil {
		return 0, 0, err
	}
	return int64(ints[0]), int64(ints[1]), nil
}

func lt_fun(args []MalType) (MalType, error) {
	n1, n2, err := get_two_ints_for_cmp(args)
	if err != nil {
		return nil, err
	}
	return MalTypeBoolean(n1 < n2), nil
}

func le_fun(args []MalType) (MalType, error) {
	n1, n2, err := get_two_ints_for_cmp(args)
	if err != nil {
		return nil, err
	}
	return MalTypeBoolean(n1 <= n2), nil
}

func gt_fun(args []MalType) (MalType, error) {
	n1, n2, err := get_two_ints_for_cmp(args)
	if err != nil {
		return nil, err
	}
	return MalTypeBoolean(n1 > n2), nil
}

func ge_fun(args []MalType) (MalType, error) {
	n1, n2, err := get_two_ints_for_cmp(args)
	if err != nil {
		return nil, err
	}
	return MalTypeBoolean(n1 >= n2), nil
}

func read_string_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 1, 1)
	if err != nil {
		return nil, err
	}
	arg_str, err := assert_string_arg(args[0])
	if err != nil {
		return nil, err
	}
	return ReadStr(string(arg_str))
}

func slurp_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 1, 1)
	if err != nil {
		return nil, err
	}
	arg_str, err := assert_string_arg(args[0])
	if err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(string(arg_str))
	if err != nil {
		return nil, err
	}

	return MalTypeString(string(contents)), nil
}

func atom_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 1, 1)
	if err != nil {
		return nil, err
	}
	arg := args[0]
	return MalTypeAtom(&arg), nil
}

func get_arg_as_atom(args []MalType) (MalTypeAtom, error) {
	err := assert_len_args(args, 1, 1)
	if err != nil {
		return nil, err
	}
	arg_atom, err := assert_atom_arg(args[0])
	if err != nil {
		return nil, err
	}
	return arg_atom, nil
}
	

func atom_pred_fun(args []MalType) (MalType, error) {
	_, err := get_arg_as_atom(args)
	return MalTypeBoolean(err == nil), nil
}

func deref_fun(args []MalType) (MalType, error) {
	atom, err := get_arg_as_atom(args)
	if err != nil {
		return nil, err
	}
	return *atom, nil
}

func reset_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 2, 2)
	if err != nil {
		return nil, err
	}
	atom, err := assert_atom_arg(args[0])
	if err != nil {
		return nil, err
	}
	*atom = args[1]
	return args[1], nil
}

func swap_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 2, -1)
	if err != nil {
		return nil, err
	}
	atom, err := assert_atom_arg(args[0])
	if err != nil {
		return nil, err
	}

	fun, err := assert_function_arg(args[1])
	if err != nil {
		return nil, err
	}

	new_args := make([]MalType, len(args)-1)
	new_args[0] = *atom
	for i := 0; i<len(args)-2; i++ {
		new_args[i+1] = args[i+2]
	}
	res, err := fun(new_args)
	if err != nil {
		return nil, err
	}
	*atom = res
	return res, nil
}

func cons_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 2, -1)
	if err != nil {
		return nil, err
	}
	arr, err := assert_list_or_vec_arg(args[1])
	if err != nil {
		return nil, err
	}

	res := make([]MalType, len(arr)+1)
	copy(res[1:], arr)
	res[0] = args[0]
	return MalTypeList(res), nil
}

func concat_fun(args []MalType) (MalType, error) {
	args_lists, err := assert_list_or_vec_args(args)
	if err != nil {
		return nil, err
	}

	n := 0
	for _, lst := range(args_lists) {
		n += len(lst)
	}

	res := make([]MalType, n)
	n = 0
	for _, lst := range(args_lists) {
		for _, x := range(lst) {
			res[n] = x
			n += 1
		}
	}

	return MalTypeList(res), nil
}

func vec_fun(args []MalType) (MalType, error) {
	err := assert_len_args(args, 1, 1)
	if err != nil {
		return nil, err
	}
	arr, err := assert_list_or_vec_arg(args[0])
	if err != nil {
		return nil, err
	}

	res := make([]MalType, len(arr))
	copy(res, arr)
	return MalTypeVector(res), nil
}

func Ns() map[string]func([]MalType)(MalType, error) {
	res := make(map[string]func([]MalType)(MalType, error))
	res["+"] = sum_fun
	res["-"] = sub_fun
	res["*"] = mul_fun
	res["/"] = div_fun
	res["prn"] = prn_fun
	res["list"] = list_fun
	res["list?"] = list_pred_fun
	res["empty?"] = empty_pred_fun
	res["count"] = count_fun
	res["="] = eq_fun
	res["<"] = lt_fun
	res["<="] = le_fun
	res[">"] = gt_fun
	res[">="] = ge_fun
	res["pr-str"] = pr_str_fun
	res["str"] = str_fun
	res["println"] = println_fun
	res["read-string"] = read_string_fun
	res["slurp"] = slurp_fun
	res["atom"] = atom_fun
	res["atom?"] = atom_pred_fun
	res["deref"] = deref_fun
	res["reset!"] = reset_fun
	res["swap!"] = swap_fun
	res["cons"] = cons_fun
	res["concat"] = concat_fun
	res["vec"] = vec_fun
	return res
}
