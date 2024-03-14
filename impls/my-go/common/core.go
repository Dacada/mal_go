package common

import (
	"errors"
	"fmt"
	"strings"
)

func assert_int_args(args []MalType) error {
	for i, element := range args {
		res, ok := element.(MalTypeInteger)
		if !ok {
			return errors.New(fmt.Sprintf("argument %d expected to be of type %T but was %T", i, res, element))
		}
	}
	return nil
}

func assert_nonempty_args(args []MalType) error {
	if len(args) == 0 {
		return errors.New("unexpected number of arguments for builtin function, expected at least 1")
	}
	return nil
}

func sum_fun(args []MalType) (MalType, error) {
	err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(0)
	for _, element := range args {
		int_element := element.(MalTypeInteger)
		res += int64(int_element)
	}

	return MalTypeInteger(res), nil
}

func sub_fun(args []MalType) (MalType, error) {
	if len(args) == 0 {
		return MalTypeInteger(0), nil
	}
	err := assert_int_args(args)
	
	sum_res, err := sum_fun(args[1:])
	if err != nil {
		return nil, err
	}

	return MalTypeInteger(args[0].(MalTypeInteger) - sum_res.(MalTypeInteger)), nil
}

func mul_fun(args []MalType) (MalType, error) {
	err := assert_int_args(args)
	if err != nil {
		return nil, err
	}

	res := int64(1)
	for _, element := range args {
		int_element := element.(MalTypeInteger)
		res *= int64(int_element)
	}

	return MalTypeInteger(res), nil
}

func div_fun(args []MalType) (MalType, error) {
	err := assert_nonempty_args(args)
	if err != nil {
		return nil, err
	}
	err = assert_int_args(args)
	if err != nil {
		return nil, err
	}

	if len(args) == 1 {
		return MalTypeInteger(1 / int64(args[0].(MalTypeInteger))), nil
	}

	rest, err := mul_fun(args[1:])
	if err != nil {
		return nil, err
	}
	return MalTypeInteger(args[0].(MalTypeInteger) / rest.(MalTypeInteger)), nil
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
	println(string(str.(MalTypeString)))
	return MalTypeNil{}, nil
}

func println_fun(args []MalType) (MalType, error) {
	strs := make([]string, len(args))
	for i, arg := range args {
		strs[i] = PrStr(arg, false)
	}
	println(strings.Join(strs, " "))
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
	if len(args) < 2 {
		return nil, errors.New("expected at least two arguments")
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
	if len(args) < 2 {
		return 0, 0, errors.New("expected at least two arguments")
	}
	n1, ok := args[0].(MalTypeInteger)
	if !ok {
		return 0, 0, errors.New(fmt.Sprintf("expected first argument to be %T but is %T", n1, args[0]))
	}
	n2, ok := args[1].(MalTypeInteger)
	if !ok {
		return 0, 0, errors.New(fmt.Sprintf("expected second argument to be %T but is %T", n2, args[1]))
	}
	return int64(n1), int64(n2), nil
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
	return res
}
