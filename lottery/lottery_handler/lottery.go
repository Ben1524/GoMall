package lottery_handler

import "math/rand"

const (
	NoneItem = -1
)

func Lottery(probs []float64) int {
	if len(probs) == 0 {
		return NoneItem
	}
	// 归一化
	var sum float64
	accumulate := make([]float64, 0, len(probs)) // 累积概率
	for _, p := range probs {
		sum += p
		accumulate = append(accumulate, sum)
	}

	if sum == 0 {
		return NoneItem
	}

	r := rand.Float64() * sum
	return RangeBinarySearch(accumulate, r)
}

// 区间二分查找，返回第一个大于等于target的索引
func RangeBinarySearch(arr []float64, target float64) int {

	left, right := 0, len(arr)-1
	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	if left >= len(arr) {
		return NoneItem
	}
	return left
}
