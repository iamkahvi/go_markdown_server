package main

import (
	"log"

	"golang.org/x/exp/constraints"
)

type MyQueue[T constraints.Integer] struct {
	q   []T
	len int
}

func (mq *MyQueue[uint32]) Head() (uint32, bool) {
	if len(mq.q) == 0 {
		return uint32(0), true
	}

	log.Println("head: ", mq.q[len(mq.q)-mq.len])
	return mq.q[len(mq.q)-mq.len], false
}

func (mq *MyQueue[uint32]) Push(val uint32) {
	log.Println("pushing: ", val)
	mq.q = append(mq.q, val)
	mq.len += 1
}

func (mq *MyQueue[uint32]) Pop() (uint32, bool) {
	if len(mq.q) == 0 {
		return 0, true
	}

	var temp = mq.q[len(mq.q)-1]
	mq.q = mq.q[:len(mq.q)-1]
	mq.len -= 1

	log.Println("popping: ", temp)

	return temp, false
}

func (mq *MyQueue[uint32]) Remove(val uint32) bool {
	log.Println("removing: ", val)
	for i, v := range mq.q {
		if v == val {
			mq.q = append(mq.q[:i], mq.q[i+1:]...)
			mq.len -= 1
			return false
		}
	}

	return true
}
