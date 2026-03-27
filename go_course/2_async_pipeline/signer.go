package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan any)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	for i, j := range jobs {
		currOut := make(chan any)
		go func(in chan any, out chan any, i int) {
			defer close(out)
			if i == len(jobs)-1 {
				defer wg.Done()
			}
			j(in, out)
		}(in, currOut, i)
		in = currOut
	}
	wg.Wait()
}

// SingleHash - crc32(data)+"~"+crc32(md5(data)
func SingleHash(in chan any, out chan any) {
	wg := new(sync.WaitGroup)
	for v := range in {
		wg.Add(1)
		strV := resolveString(v)
		md5Val := DataSignerMd5(strV)
		go calcSingleHash(strV, md5Val, out, wg)
	}
	wg.Wait()
}

func calcSingleHash(v string, md5 string, out chan any, outWg *sync.WaitGroup) {
	defer outWg.Done()
	wg := new(sync.WaitGroup)
	crcCh := make(chan string, 1)
	md5Ch := make(chan string, 1)
	wg.Add(2)
	go func() {
		defer wg.Done()
		crcCh <- DataSignerCrc32(v)
	}()
	go func() {
		defer wg.Done()
		md5Ch <- DataSignerCrc32(md5)
	}()
	wg.Wait()
	out <- <-crcCh + "~" + <-md5Ch
}

func resolveString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// MultiHash crc32(th+data)), th=0..5
func MultiHash(in chan any, out chan any) {
	wg := new(sync.WaitGroup)
	for v := range in {
		wg.Add(1)
		go calcMultiHash(v.(string), out, wg)
	}
	wg.Wait()
}

func calcMultiHash(v string, out chan any, outWg *sync.WaitGroup) {
	defer outWg.Done()
	res := make([]string, 6)
	wg := new(sync.WaitGroup)
	for i := 0; i <= 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res[i] = DataSignerCrc32(strconv.Itoa(i) + v)
		}(i)
	}
	wg.Wait()
	out <- strings.Join(res, "")
}

func CombineResults(in chan any, out chan any) {
	res := make([]string, 0)
	for v := range in {
		res = append(res, v.(string))
	}
	slices.Sort(res)
	out <- strings.Join(res, "_")
}
