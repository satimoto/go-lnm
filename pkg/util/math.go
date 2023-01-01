package util

import "time"

func MaxDuration(x, y time.Duration) time.Duration {
    if x < y {
        return y
    }

    return x
}