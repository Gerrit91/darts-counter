package util

import (
	"bufio"
	"fmt"
	"os"
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

func (c *Console) Read() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
