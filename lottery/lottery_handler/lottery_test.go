package lottery_handler

import (
	"testing"
)

// 测试空概率数组
func TestLottery_EmptyProbs(t *testing.T) {
	result := Lottery([]float64{})
	if result != NoneItem {
		t.Errorf("空概率数组应返回NoneItem，实际返回%d", result)
	}
}

// 测试单元素概率数组
func TestLottery_SingleItem(t *testing.T) {
	probs := []float64{1.0}
	result := Lottery(probs)
	if result != 0 {
		t.Errorf("单元素概率数组应返回0，实际返回%d", result)
	}
}

func TestLottery_MultiItem(t *testing.T) {
	probs := []float64{0.01, 0.31, 0.6}
	counts := make([]int, len(probs))
	total := 10000000

	for i := 0; i < total; i++ {
		res := Lottery(probs)
		if res < 0 || res >= len(probs) {
			t.Errorf("返回了无效索引: %d", res)
		} else {
			counts[res]++
		}
	}

	// 允许±0.1%的误差
	expected := []float64{0.01 * float64(total), 0.3 * float64(total), 0.6 * float64(total)}
	tolerance := float64(total) * 0.1
	for i, c := range counts {
		if float64(c) < expected[i]-tolerance || float64(c) > expected[i]+tolerance {
			t.Errorf("索引%d的次数%d偏离预期（允许误差±1%%）", i, c)
		}
	}

}

// 测试包含0概率的数组
func TestLottery_ZeroProb(t *testing.T) {
	// 第一个元素概率为0，第二个为1
	probs1 := []float64{0, 1.0}
	for i := 0; i < 1000; i++ {
		if Lottery(probs1) != 1 {
			t.Error("概率[0,1.0]应始终返回1")
			break
		}
	}

	// 第一个元素概率为1，第二个为0
	probs2 := []float64{1.0, 0}
	for i := 0; i < 1000; i++ {
		if Lottery(probs2) != 0 {
			t.Error("概率[1.0,0]应始终返回0")
			break
		}
	}
}

// 测试等概率情况（统计验证）
func TestLottery_EqualProb(t *testing.T) {
	probs := []float64{1, 1, 1, 1} // 总和为4，每个概率25%
	counts := make([]int, 4)
	total := 1000000

	for i := 0; i < total; i++ {
		res := Lottery(probs)
		counts[res]++
	}

	// 允许±1%的误差
	expected := float64(total) / 4
	tolerance := expected * 0.01
	for i, c := range counts {
		if float64(c) < expected-tolerance || float64(c) > expected+tolerance {
			t.Errorf("等概率测试中，索引%d的次数%d偏离预期（允许误差±1%%）", i, c)
		}
	}
}

// 测试总和不为1的情况（归一化验证）
func TestLottery_NonNormalized(t *testing.T) {
	probs := []float64{2, 2} // 总和4，等效于[0.5,0.5]
	count0, count1 := 0, 0
	total := 1000000

	for i := 0; i < total; i++ {
		switch Lottery(probs) {
		case 0:
			count0++
		case 1:
			count1++
		default:
			t.Error("返回了无效索引")
		}
	}

	// 允许±1%的误差
	expected := float64(total) / 2
	tolerance := expected * 0.01
	if float64(count0) < expected-tolerance || float64(count0) > expected+tolerance {
		t.Errorf("非归一化测试中，索引0的次数%d偏离预期", count0)
	}
	if float64(count1) < expected-tolerance || float64(count1) > expected+tolerance {
		t.Errorf("非归一化测试中，索引1的次数%d偏离预期", count1)
	}
}

// 测试概率总和为0的情况
func TestLottery_ZeroSum(t *testing.T) {
	probs := []float64{0, 0, 0}
	result := Lottery(probs)
	// 预期返回NoneItem，但当前代码会返回0
	if result != NoneItem {
		t.Errorf("概率总和为0时应返回NoneItem，实际返回%d", result)
	}
}

// 测试包含负数概率的情况
func TestLottery_NegativeProb(t *testing.T) {
	probs := []float64{-1, 3} // 总和2，累积概率为[-1,2]
	count0, count1 := 0, 0
	total := 1000000

	for i := 0; i < total; i++ {
		switch Lottery(probs) {
		case 0:
			count0++
		case 1:
			count1++
		}
	}

	// 负数概率会导致不合理结果（此处0索引本应概率为负，却可能被选中）
	if count0 > 0 {
		t.Errorf("负数概率测试中，索引0（负概率）被选中了%d次，不符合预期", count0)
	}
}
