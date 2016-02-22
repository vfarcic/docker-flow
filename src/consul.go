package main
import (
	"strconv"
)

func getConsulScale(serviceName, scale string) int {
	// TODO: Get from Consul
	s := 3
	inc := 0
	if len(scale) > 0 {
		if scale[:1] == "+" || scale[:1] == "-" {
			inc, _ = strconv.Atoi(scale)
		} else {
			s, _ = strconv.Atoi(scale)
		}
	}
	total := s + inc
	if total <= 0 {
		return 1
	}
	return total
}
