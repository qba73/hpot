package main

import "github.com/qba73/hpot"

func main() {
	pot := hpot.NewHoneyPotServer()
	pot.Ports = []int{8081, 8082}

	if err := pot.ListenAndServe(); err != nil {
		panic(err)
	}
}
