package repository

import (
	"fmt"
	"log/slog"
	"product/domain/model"

	"github.com/jinzhu/gorm"
)
type IProductRepository interface{
    InitTable() error
    FindProductByID(int64) (*model.Product, error)
	CreateProduct(*model.Product) (int64, error)
	DeleteManyProductByIDs(...int64) error
	DeleteProductByID(int64) error
	UpdateProduct(*model.Product) error
	FindAll()([]model.Product,error)

}
//创建productRepository
func NewProductRepository(db *gorm.DB) IProductRepository  {
	return &ProductRepository{mysqlDb:db}
}

type ProductRepository struct {
	mysqlDb *gorm.DB
}

//初始化表
func (u *ProductRepository)InitTable() error  {
	return u.mysqlDb.CreateTable(&model.Product{},&model.ProductSeo{},&model.ProductImage{},&model.ProductSize{}).Error
}

//根据ID查找Product信息
func (u *ProductRepository)FindProductByID(productID int64) (product *model.Product,err error) {
	product = &model.Product{}
	return product, u.mysqlDb.Preload("ProductImage").Preload("ProductSize").Preload("ProductSeo").First(product,productID).Error
}

//创建Product信息
func (u *ProductRepository) CreateProduct(product *model.Product) (int64, error) {
	return product.ID, u.mysqlDb.Create(product).Error
}

func (u *ProductRepository) DeleteManyProductByIDs(productIDs ...int64) error {
	// 开启事务
	tx := u.mysqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("删除产品时发生panic,[productIDs=", productIDs, "]", r)
		}
	}()

	if tx.Error != nil {
		slog.Error("开启事务失败", "productIDs", productIDs, "error", tx.Error.Error())
		return tx.Error
	}

	// 辅助函数：执行删除并处理错误
	deleteWithTx := func(where string, dest interface{}) error {
		if err := tx.Unscoped().Where(where, productIDs).Delete(dest).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除失败: %w", err)
		}
		return nil
	}
	// 1. 先删除关联表数据（顺序：子表 -> 主表）
	if err := deleteWithTx("images_product_id IN (?)", &model.ProductImage{}); err != nil {
		slog.Error("删除产品图片失败", "productIDs", productIDs, "error", err.Error())
		return err
	}
	
	if err := deleteWithTx("size_product_id IN (?)", &model.ProductSize{}); err != nil {
		slog.Error("删除产品尺寸失败", "productIDs", productIDs, "error", err.Error())
		return err
	}
	if err := deleteWithTx("seo_product_id IN (?)", &model.ProductSeo{}); err != nil {
		slog.Error("删除产品SEO信息失败", "productIDs", productIDs, "error", err.Error())
		return err
	}
	
	// 2. 最后删除主表产品
	if err := deleteWithTx("id IN (?)", &model.Product{}); err != nil {
		slog.Error("删除产品主表失败", "productIDs", productIDs, "error", err.Error())
		return err
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		slog.Error("事务提交失败", "productIDs", productIDs, "error", err.Error())
		return err
	}

	slog.Info("产品及关联数据删除成功", "productIDs", productIDs)
	return nil
}

// 根据ID删除Product信息（包含关联数据）
func (u *ProductRepository) DeleteProductByID(productID int64) error {
	// 开启事务
	tx := u.mysqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			slog.Error("删除产品时发生panic,[productID=", productID, "]", r)
		}
	}()

	if tx.Error != nil {
		slog.Error("开启事务失败", slog.Int64("productID", productID), slog.String("error", tx.Error.Error()))
		return tx.Error
	}

	// 辅助函数：执行删除并处理错误
	deleteWithTx := func(where string, dest interface{}) error {
		if err := tx.Unscoped().Where(where, productID).Delete(dest).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除失败: %w", err)
		}
		return nil
	}

	// 1. 先删除关联表数据（顺序：子表 -> 主表）
	if err := deleteWithTx("images_product_id = ?", &model.ProductImage{}); err != nil {
		slog.Error("删除产品图片失败", slog.Int64("productID", productID), slog.String("error", err.Error()))
		return err
	}

	if err := deleteWithTx("size_product_id = ?", &model.ProductSize{}); err != nil {
		slog.Error("删除产品尺寸失败", slog.Int64("productID", productID), slog.String("error", err.Error()))
		return err
	}

	if err := deleteWithTx("seo_product_id = ?", &model.ProductSeo{}); err != nil {
		slog.Error("删除产品SEO信息失败", slog.Int64("productID", productID), slog.String("error", err.Error()))
		return err
	}

	// 2. 最后删除主表产品
	if err := deleteWithTx("id = ?", &model.Product{}); err != nil {
		slog.Error("删除产品主表失败", slog.Int64("productID", productID), slog.String("error", err.Error()))
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		slog.Error("事务提交失败", slog.Int64("productID", productID), slog.String("error", err.Error()))
		return err
	}

	slog.Info("产品及关联数据删除成功", slog.Int64("productID", productID))
	return nil
}

//更新Product信息
func (u *ProductRepository) UpdateProduct(product *model.Product) error {
	return u.mysqlDb.Model(product).Update(product).Error
}

//获取结果集
func (u *ProductRepository) FindAll()(productAll []model.Product,err error) {
	return productAll, u.mysqlDb.Preload("ProductImage").Preload("ProductSize").Preload("ProductSeo").Find(&productAll).Error
}

