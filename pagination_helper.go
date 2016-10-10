package main

type PaginationHelper struct {
	Page     int
	IsActive bool
}

func CreatePaginationHelper(current, max, choices int) []PaginationHelper {
	left := current - choices
	right := current + choices

	if left <= 0 {
		left = 1
	}

	if right > max {
		right = max
	}

	arr := make([]PaginationHelper, 0, right-left+1+4)

	if left != 1 {
		arr = append(arr, PaginationHelper{1, false})
		arr = append(arr, PaginationHelper{-1, false})
	}

	for i := left; i <= right; i++ {
		isActive := false

		if i == current {
			isActive = true
		}

		arr = append(arr, PaginationHelper{i, isActive})
	}

	if right != max {
		arr = append(arr, PaginationHelper{-1, false})
		arr = append(arr, PaginationHelper{max, false})
	}

	return arr
}
