package util

import (
    "log"
    "strconv"
)

func Str2float64(str string) float64 {
    value, err := strconv.ParseFloat(str, 64)
    if err != nil {
        log.Printf("Str2float64: failed to parse '%s': %v", str, err)
        return 0
    }
    return value
}
