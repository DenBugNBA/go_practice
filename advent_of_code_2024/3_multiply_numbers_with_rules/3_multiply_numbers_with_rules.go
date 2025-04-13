package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	disableOperandFlag = "don't()"
	enableOperandFlag  = "do()"
)

func main() {
	lines, err := readFile("./advent_of_code_2024/input_data/3/input.txt")
	if err != nil {
		log.Error().Err(err).Msg("on reading file")
		return
	}
	sum := 0
	mulEnabled := true
	for _, l := range lines {
		matchingStrings := findMatchingStrings(l)
		for _, match := range matchingStrings {
			switch match[0] {
			case disableOperandFlag:
				mulEnabled = false
			case enableOperandFlag:
				mulEnabled = true
			default:
				if mulEnabled {
					currSum, err := multiply(match)
					if err != nil {
						log.Error().Err(err).Msg("on multiplying")
					}
					sum += currSum
				}
			}
		}
	}
	fmt.Println(sum)
}

func readFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Error().Err(err).Msg("on closing file")
		}
	}()
	s := bufio.NewScanner(file)
	lines := make([]string, 0)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines, nil
}

func findMatchingStrings(text string) [][]string {
	reg := regexp.MustCompile("mul\\((\\d+),(\\d+)\\)|do\\(\\)|don't\\(\\)")
	return reg.FindAllStringSubmatch(text, -1)
}

func multiply(match []string) (int, error) {
	num1, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	num2, err := strconv.Atoi(match[2])
	if err != nil {
		return 0, err
	}
	return num1 * num2, nil
}
