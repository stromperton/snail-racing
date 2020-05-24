package main

import (
	"math/rand"
	"time"
)

//UpdateEvery Обновление...
func UpdateEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		f()
	}

}

//Random Случайное чилсо от min до max [min; max)
func Random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
