package main

import "log"

type Queue[T uint32 | bool] interface {
	Head() T
	Push(T)
	Pop() (T, bool)
	Remove(T)
}

type MyQueue struct {
	q []uint32
}

type Value interface {
	// Somehow VScode won't let me define generics
	uint32 | bool
}

func (mq *MyQueue) Head() (uint32, bool) {
	if len(mq.q) == 0 {
		return 0, false
	}

	return mq.q[len(mq.q)-1], true
}

func (mq *MyQueue) Push(val uint32) {
	log.Println("pushing: ", val)
	mq.q = append(mq.q, val)
}

func (mq *MyQueue) Pop() (uint32, bool) {
	if len(mq.q) == 0 {
		return 0, false
	}

	var temp = mq.q[len(mq.q)-1]
	mq.q = mq.q[:len(mq.q)-1]

	log.Println("popping: ", temp)

	return temp, true
}

func (mq *MyQueue) Remove(val uint32) bool {
	log.Println("removing: ", val)
	for i, v := range mq.q {
		if v == val {
			mq.q = append(mq.q[:i], mq.q[i+1:]...)
			return true
		}
	}

	return false
}
