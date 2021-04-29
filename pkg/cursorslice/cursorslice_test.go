package cursorslice

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var items = func() []interface{} {
	const numItems = 1000000
	items := make([]interface{}, numItems, numItems)
	for i := range items {
		items[i] = i
	}
	return items
}()

func TestCS(t *testing.T) {
	cs := NewCursorSlice()
	cs.Append(items...)

	var wg sync.WaitGroup
	c := make(chan []interface{})
	f := func(s string) func(key int, value interface{}) bool {
		wg.Add(1)
		return func(key int, value interface{}) bool {
			c <- []interface{}{s, key, value}
			return true
		}
	}

	winner := "x"
	a := f("a")
	go func() {
		defer wg.Done()
		cs.Range(a)
		winner = "a"
	}()

	b := f("b")
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		cs.Range(b)
		winner = "b"
	}()

	go func() {
		for pr := range c {
			_ = pr
			//fmt.Println(pr...)
		}
		fmt.Println(winner)
	}()
	wg.Wait()
	close(c)

}

func TestMassive(t *testing.T) {
	cs := NewCursorSlice()
	cs.Append(items...)

	var wg sync.WaitGroup
	c := make(chan []interface{})
	f := func(s interface{}) func(key int, value interface{}) bool {
		wg.Add(1)
		return func(key int, value interface{}) bool {
			c <- []interface{}{s, key, value}
			return true
		}
	}

	for i := 0; i < 9; i++ {
		a := f(i)
		go func() {
			defer wg.Done()
			cs.Range(a)
		}()
	}

	go func() {
		for pr := range c {
			_ = pr
			//fmt.Println(pr...)
		}
	}()

	wg.Wait()
	close(c)

}
