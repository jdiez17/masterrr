package main

import (
	"fmt"
)

func key(key string) string {
	return fmt.Sprintf("masterrr.%d.%s", State.ID, key)
}
