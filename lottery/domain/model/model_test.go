package model

import (
	"testing"
	"time"

	"lottery/domain/model/product"

	config "github.com/Ben1524/GoMall/common/config"
	db "github.com/Ben1524/GoMall/common/db"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局测试数据库连接
var testDB *gorm.DB

// 初始化测试数据库（使用SQLite内存库）
func init() {
	var err error
	cfg, err := config.Load("D:\\GolandProjects\\GoMall\\lottery\\config.example.yaml")
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}

	testDB, err = db.NewMysqlGorm(cfg)

	if err != nil {
		panic("连接测试数据库失败: " + err.Error())
	}

	testDB.AutoMigrate(&LotteryAward{}, &LotteryActivity{})

}

// 测试2: 唯一约束有效性验证
func TestUniqueConstraints(t *testing.T) {
	// 准备测试数据
	product := product.Product{ProductName: "测试商品", ProductSku: "TEST_SKU_001"}
	//testDB.Create(&product)

	activity := LotteryActivity{
		ActivityName: "测试活动",
		StartTime:    time.Now().Add(1 * time.Hour),
		EndTime:      time.Now().Add(2 * time.Hour),
		Status:       0,
	}
	testDB.Create(&activity)

	// 测试1: 同一活动添加相同产品+规格的奖品（应触发唯一约束）
	award1 := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     product.ID,
		ProductSizeID: nil,
		Probability:   10,
		TotalQuantity: 100,
	}
	if err := testDB.Create(&award1).Error; err != nil {
		t.Fatal("创建第一个奖品失败: " + err.Error())
	}

	// 重复添加相同奖品
	award2 := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     product.ID,
		ProductSizeID: nil,
		Probability:   20,
		TotalQuantity: 50,
	}
	if err := testDB.Create(&award2).Error; err == nil {
		t.Error("同一活动添加相同产品+规格的奖品未触发唯一约束，不符合预期")
	}

	// 测试2: 同一产品在同一时间范围创建重复活动（应触发唯一约束）
	duplicateActivity := LotteryActivity{
		ActivityName: "重复活动",
		StartTime:    activity.StartTime,
		EndTime:      activity.EndTime,
		Status:       0,
	}
	if err := testDB.Create(&duplicateActivity).Error; err == nil {
		t.Error("同一产品在同一时间范围创建重复活动未触发唯一约束，不符合预期")
	}
}

// 测试3: 外键约束验证
func TestForeignKeyConstraints(t *testing.T) {
	// 准备测试数据
	activity := LotteryActivity{
		ActivityName: "外键测试活动",
		StartTime:    time.Now().Add(1 * time.Hour),
		EndTime:      time.Now().Add(2 * time.Hour),
		Status:       0,
	}
	testDB.Create(&activity)

	// 测试1: 关联不存在的产品ID（应失败）
	invalidAward := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     999999, // 不存在的产品ID
		Probability:   10,
		TotalQuantity: 100,
	}
	if err := testDB.Create(&invalidAward).Error; err == nil {
		t.Error("关联不存在的产品ID未触发外键约束，不符合预期")
	}

	// 测试2: 删除产品后检查外键行为
	product := product.Product{ProductName: "待删除商品", ProductSku: "TEST_SKU_002"}
	testDB.Create(&product)

	award := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     product.ID,
		Probability:   10,
		TotalQuantity: 100,
	}
	testDB.Create(&award)

	// 删除产品
	testDB.Delete(&product)

	// 检查关联的奖品是否仍存在（外键允许存在，仅product_id无效）
	var count int64
	testDB.Model(&LotteryAward{}).Where("id = ?", award.ID).Count(&count)
	if count == 0 {
		t.Error("删除关联产品后，奖品记录被意外删除，不符合外键约束设计")
	}
}

// 测试4: 业务场景合理性验证
func TestBusinessScenarios(t *testing.T) {
	// 场景1: 创建一个活动并添加多个不同奖品
	activity := LotteryActivity{
		ActivityName: "多奖品测试活动",
		StartTime:    time.Now().Add(1 * time.Hour),
		EndTime:      time.Now().Add(2 * time.Hour),
		Status:       0,
	}
	testDB.Create(&activity)

	// 创建两个不同产品
	product1 := product.Product{ProductName: "奖品A", ProductSku: "PRIZE_A_001"}
	product2 := product.Product{ProductName: "奖品B", ProductSku: "PRIZE_B_001"}
	testDB.Create(&product1)
	testDB.Create(&product2)

	// 添加两个不同奖品
	award1 := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     product1.ID,
		Probability:   30, // 30%概率
		TotalQuantity: 50,
	}
	award2 := LotteryAward{
		ActivityID:    activity.ID,
		ProductID:     product2.ID,
		Probability:   70, // 70%概率
		TotalQuantity: 100,
	}
	testDB.Create(&award1)
	testDB.Create(&award2)

	// 验证奖品总数
	var totalAwards int64
	testDB.Model(&LotteryAward{}).Where("activity_id = ?", activity.ID).Count(&totalAwards)
	if totalAwards != 2 {
		t.Errorf("多奖品场景下，预期2个奖品，实际%d个", totalAwards)
	}

	// 场景2: 验证概率总和是否合理（业务逻辑检查）
	var sumProbability float64
	testDB.Model(&LotteryAward{}).Where("activity_id = ?", activity.ID).Select("SUM(probability)").Scan(&sumProbability)
	if sumProbability > 100 {
		t.Errorf("奖品概率总和超过100%%，实际%.2f%%，可能导致业务逻辑异常", sumProbability)
	}

	// 场景3: 验证库存减少逻辑
	var award LotteryAward
	testDB.First(&award, "activity_id = ? AND product_id = ?", activity.ID, product1.ID)
	initialRemaining := award.RemainingQuantity

	// 模拟抽奖后库存减少
	testDB.Model(&award).Update("remaining_quantity", gorm.Expr("remaining_quantity - 1"))
	testDB.First(&award, award.ID) // 重新查询

	if award.RemainingQuantity != initialRemaining-1 {
		t.Errorf("库存减少逻辑异常，预期%d，实际%d", initialRemaining-1, award.RemainingQuantity)
	}
}
