package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
)

func main() {
	lines, err := readFile("./advent_of_code_2024/input_data/3/input.txt")
	if err != nil {
		log.Error().Err(err).Msg("on reading file")
		return
	}
	sum := 0
	for _, l := range lines {
		matchingStrings := findMatchingStrings(l)
		lineSum, err := sumStrings(matchingStrings)
		if err != nil {
			log.Error().Err(err).Msg("on calculating sum")
			return
		}
		sum += lineSum
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

func findMatchingStrings(text string) []string {
	reg := regexp.MustCompile("mul\\(\\d+,\\d+\\)")
	return reg.FindAllString(text, -1)
}

func sumStrings(strings []string) (int, error) {
	sum := 0
	for _, s := range strings {
		currMultiply, err := multiply(s)
		if err != nil {
			return 0, err
		}
		sum += currMultiply
	}
	return sum, nil
}

func multiply(s string) (int, error) {
	numbersReg := regexp.MustCompile("\\d+")
	numbers := numbersReg.FindAllString(s, -1)
	num1, err := strconv.Atoi(numbers[0])
	if err != nil {
		return 0, err
	}
	num2, err := strconv.Atoi(numbers[1])
	if err != nil {
		return 0, err
	}
	return num1 * num2, nil
}
