package utils

import (
	"fmt"
	"strings"
)

type Int2 struct {
	X, Y int
}

func NewInt2(x, y int) Int2 {
	return Int2{x, y}
}

func (i Int2) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("(%d,%d)", i.X, i.Y)), nil
}

func (i *Int2) UnmarshalText(text []byte) error {
	_, err := fmt.Fscanf(strings.NewReader(string(text)), "(%d,%d)", &i.X, &i.Y)
	return err
}

type Float2 struct {
	X, Y float32
}

func NewFloat2(x, y float32) Float2 {
	return Float2{x, y}
}
