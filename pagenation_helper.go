package main

type PagenationHelper struct {
    Page int
    IsActive bool
}

func CreatePagenationHelper(current, max, choices int) []PagenationHelper {
    left := current - choices
    right := current + choices

    if left <= 0 {
        left = 1
    }

    if right > max {
        right = max
    }

    arr := make([]PagenationHelper, 0, right - left + 1 + 4)

    if left != 1 {
        arr = append(arr, PagenationHelper{1, false})
        arr = append(arr, PagenationHelper{-1, false})
    }

    for i := left; i <= right; i++ {
        isActive := false

        if i == current {
            isActive = true
        }

        arr = append(arr, PagenationHelper{i, isActive})
    }

    if right != max {
        arr = append(arr, PagenationHelper{-1, false})
        arr = append(arr, PagenationHelper{max, false})
    }

    return arr
}