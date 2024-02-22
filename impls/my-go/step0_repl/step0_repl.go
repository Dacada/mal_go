package main;

import (
	"fmt"
	"bufio"
	"os"
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
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("user> ")
		if !scanner.Scan() {
			err := scanner.Err();
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
			break
		}
		fmt.Println(rep(scanner.Text()))
	}
}
