package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	return rand.Intn(max-min) + min
}

func GetInt(key string) int {
	num, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		fmt.Println("Проблема с парсом переменной окружения "+key, err)
	}
	return num

}
