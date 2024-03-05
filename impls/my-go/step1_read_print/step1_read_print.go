package main;

import (
	"os"
	"fmt"
	"example.com/mal/common"
	"github.com/chzyer/readline"
)

func READ(input string) (common.MalType, error) {
	return common.ReadStr(input)
}

func EVAL(input common.MalType) (common.MalType, error) {
	return input, nil
}

func PRINT(input common.MalType) string {
	return common.PrStr(input, true)
}

func rep(input string) (string, error) {
	read_val, err := READ(input)
	if err != nil {
		return "", err
	}
	eval_val, err := EVAL(read_val)
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

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		
		ret, err := rep(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, err);
		} else {
			fmt.Println(ret)
		}
	}
}
