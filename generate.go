package main

import (
	"fmt"
	"slices"
)

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

func main1() {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{6, 7, 8, 9, 10}

	newSlice := slices.AppendSeq(slice1, slices.Values(slice2))

	fmt.Println(newSlice)
}

func main() {
	s := []int{1, 2, 3, 4, 5}

	fmt.Println(slices.Collect(slices.Values(s)))
}
