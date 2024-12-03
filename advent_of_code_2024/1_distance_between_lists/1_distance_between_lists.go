package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("./advent_of_code_2024/input_data/1/input.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()

	list1, list2 := make([]int, 0), make([]int, 0)

	s := bufio.NewScanner(file)

	for s.Scan() {
		line := s.Text()
		nums := strings.Fields(line)
		num1, err := strconv.Atoi(nums[0])
		if err != nil {
			panic(err)
		}
		list1 = append(list1, num1)
		num2, err := strconv.Atoi(nums[1])
		if err != nil {
			panic(err)
		}
		list2 = append(list2, num2)
	}

	sort.Ints(list1)
	sort.Ints(list2)

	distance := 0
	for i := 0; i < len(list1); i++ {
		distance += int(math.Abs(float64(list1[i] - list2[i])))
	}
	fmt.Println(distance)
}
