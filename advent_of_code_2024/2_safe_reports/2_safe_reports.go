package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("./advent_of_code_2024/input_data/2/input.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	s := bufio.NewScanner(file)

	safeCount := 0
	for s.Scan() {
		if areSafeLevels(parseLevels(s.Text())) {
			safeCount++
		}
	}
	fmt.Println(safeCount)
}

func parseLevels(line string) []int {
	parts := strings.Fields(line)
	numbers := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			panic(fmt.Sprintf("unexpected char in line %s: %s", line, err.Error()))
		}
		numbers = append(numbers, n)
	}
	return numbers
}

type levelsType int

const (
	increasing levelsType = iota
	decreasing
	undefined
	invalid
)

func areSafeLevels(levels []int) bool {
	currLevelsType := defineLevelsType(levels)
	switch currLevelsType {
	case increasing:
		return areSafeIncreasingLevels(levels)
	case decreasing:
		return areSafeDecreasingLevels(levels)
	case undefined:
		return true
	case invalid:
		return false
	default:
		return false
	}
}

func defineLevelsType(levels []int) levelsType {
	if len(levels) == 1 {
		return undefined
	}
	if levels[0] == levels[1] {
		return invalid
	}
	if levels[0] < levels[1] {
		return increasing
	}
	return decreasing
}

func areSafeIncreasingLevels(levels []int) bool {
	prevLevel := levels[0]
	for i := 1; i < len(levels); i++ {
		if !isValidPair(prevLevel, levels[i]) || prevLevel >= levels[i] {
			return false
		}
		prevLevel = levels[i]
	}
	return true
}

func areSafeDecreasingLevels(levels []int) bool {
	prevLevel := levels[0]
	for i := 1; i < len(levels); i++ {
		if !isValidPair(prevLevel, levels[i]) || prevLevel <= levels[i] {
			return false
		}
		prevLevel = levels[i]
	}
	return true
}

func isValidPair(num1, num2 int) bool {
	diff := int(math.Abs(float64(num1 - num2)))
	return diff != 0 && diff <= 3
}
