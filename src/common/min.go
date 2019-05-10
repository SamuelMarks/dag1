package common

// Min function implement minimum for unlimited number of integer arguments
func Min(nums... int) int {
	var min int
	for i, v := range nums {
		if i == 0 || v < min {
			min = v
		}
	}
	return min
}
