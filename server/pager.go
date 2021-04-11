package server

import (
	"strconv"
)

const PAGE_SIZE = 20
const PAGER_SIDE_SIZE = 5

type pager struct {
	First   string
	Prev    []string
	Current string
	Next    []string
	Last    string
}

func createPager(current, last, sideSize int) pager {
	var p pager
	if last > 1 {
		p.Current = strconv.Itoa(current)
	}
	if current <= sideSize+2 {
		for i := 1; i < current; i += 1 {
			p.Prev = append(p.Prev, strconv.Itoa(i))
		}
	} else {
		p.First = "1"
		for i := current - sideSize; i < current; i += 1 {
			p.Prev = append(p.Prev, strconv.Itoa(i))
		}
	}
	if current >= last-(sideSize+1) {
		for i := current + 1; i <= last; i += 1 {
			p.Next = append(p.Next, strconv.Itoa(i))
		}
	} else {
		p.Last = strconv.Itoa(last)
		for i := current + 1; i <= current+sideSize; i += 1 {
			p.Next = append(p.Next, strconv.Itoa(i))
		}
	}
	return p
}
