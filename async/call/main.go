package call

import (
	"fmt"
	"time"
)

func Add(args ...int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}

func Multiply(args ...int64) (int64, error) {
	sum := int64(1)
	for _, arg := range args {
		sum *= arg
	}
	return sum, nil
}

func Cronjob() error {
	now := time.Now()
	fmt.Println(now.Format("2006=01-02 15:04:05"))
	return nil
}
