package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Console struct {
}

func (c *Console) Println(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

func (c *Console) Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

func (c *Console) AskForScore() int {
	c.Printf("enter score: ")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		score, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			c.Println("unable to parse input (%q), please enter again", err.Error())
			continue
		}

		return score
	}
}
