package main;

import (
	"fmt"
	"github.com/chzyer/readline"
)

func READ(input string) string {
	return input
}

func EVAL(input string) string {
	return input
}

func PRINT(input string) string {
	return input
}

func rep(input string) string {
	return PRINT(EVAL(READ(input)))
}

func main() {
	rl, err := readline.New("user> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		fmt.Println(rep(line))
	}
}
