# GoMall

微服务电商系统

## 配置系统（common/config）

`common/config` 包提供 YAML + 环境变量的配置加载能力，优先级为：内置默认值 < 配置文件 < 环境变量。核心特性：

- 自动寻找配置文件：依次尝试显式传入的路径、`CONFIG_FILE` 环境变量以及 `config/config.yaml` / `config.yaml` 等常用位置。
- 环境变量覆盖：字段名会被转换为大写下划线，例如 `SERVER_PORT`、`DATABASE_PASSWORD`、`JWT_SECRET`。
- 丰富的默认值与便捷方法：开箱即用，并提供 `GetDatabaseDSN`、`GetRedisAddr`、`GetRabbitMQURL` 等辅助函数。

示例文件见 `common/config/config.example.yaml`，复制为 `config/config.yaml` 后按需修改即可。

### 快速上手

```go
package main

import (
    "log"

    "GoMall/common/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("load config: %v", err)
    }

    log.Printf("server listen on %s:%s", cfg.Server.Host, cfg.Server.Port)
}
```

若希望直接在出错时终止，可使用 `config.MustLoad()`。

### 指定配置路径

```go
cfg := config.MustLoad("./deploy/config.prod.yaml")
```

### 环境变量覆盖示例

```bash
$env:SERVER_PORT="9090"
$env:JWT_SECRET="super-secret-key"
```

启动服务时，将以上环境变量应用即可覆盖对应配置项。



---

### 一、各表功能与字段详解
以下是10张表的核心作用及字段说明，覆盖电商系统的商品管理、用户管理、购物车、订单、支付、秒杀等核心模块：


#### 1. `products`（商品主表）
**核心作用**：存储商品的基础信息，是商品相关表的"主表"。  
| 字段名                | 类型         | 说明                                                         |
| --------------------- | ------------ | ------------------------------------------------------------ |
| `id`                  | bigint       | 主键（自增），唯一标识一个商品                               |
| `product_name`        | varchar(255) | 商品名称（如"2023夏季纯棉T恤"）                              |
| `product_sku`         | varchar(255) | 商品SKU编码（唯一约束），用于标识最小库存单元（如"T恤-红色-M"） |
| `product_price`       | double       | 商品原价（非秒杀价）                                         |
| `product_description` | varchar(255) | 商品描述（如材质、用途等基础信息）                           |


#### 2. `product_sizes`（商品规格表）
**核心作用**：存储商品的规格信息（如尺寸、颜色等），支持同一商品多规格。  
| 字段名            | 类型         | 说明                                                        |
| ----------------- | ------------ | ----------------------------------------------------------- |
| `id`              | bigint       | 主键（自增）                                                |
| `size_name`       | varchar(255) | 规格名称（如"S码"、"红色"）                                 |
| `size_code`       | varchar(255) | 规格唯一编码（唯一约束），如"S-RED"（避免同一商品规格重复） |
| `size_product_id` | bigint       | 关联商品ID（对应`products.id`），标识该规格属于哪个商品     |


#### 3. `product_images`（商品图片表）
**核心作用**：存储商品的图片资源（如封面图、细节图），支持商品多图展示。  
| 字段名             | 类型         | 说明                                                    |
| ------------------ | ------------ | ------------------------------------------------------- |
| `id`               | bigint       | 主键（自增）                                            |
| `image_name`       | varchar(255) | 图片名称（如"商品封面图"）                              |
| `image_code`       | varchar(255) | 图片唯一编码（唯一约束），用于区分不同图片              |
| `image_url`        | varchar(255) | 图片存储路径（如"/uploads/product/123.jpg"）            |
| `image_product_id` | bigint       | 关联商品ID（对应`products.id`），标识该图片属于哪个商品 |


#### 4. `product_seos`（商品SEO表）
**核心作用**：存储商品页面的搜索引擎优化信息，提升商品在搜索引擎中的曝光率。  
| 字段名            | 类型         | 说明                                                         |
| ----------------- | ------------ | ------------------------------------------------------------ |
| `id`              | bigint       | 主键（自增）                                                 |
| `seo_title`       | varchar(255) | 页面标题（<title>标签内容，如"2023新款T恤 - 品牌名"）        |
| `seo_keywords`    | varchar(255) | 关键词（<meta keywords>，如"T恤,新款,T恤男"）                |
| `seo_description` | varchar(255) | 描述（<meta description>，如"这款T恤采用纯棉材质，舒适透气..."） |
| `seo_code`        | varchar(255) | SEO信息编码（用于标识）                                      |
| `seo_product_id`  | bigint       | 关联商品ID（对应`products.id`），标识该SEO信息属于哪个商品   |


#### 5. `user`（用户表）
**核心作用**：存储系统用户的基础信息，是购物车、订单等功能的关联核心。  
| 字段名          | 类型         | 说明                                         |
| --------------- | ------------ | -------------------------------------------- |
| `id`            | bigint       | 主键（自增），唯一标识一个用户               |
| `username`      | varchar(50)  | 用户名（唯一约束，登录用）                   |
| `password_hash` | varchar(255) | 密码哈希（加密存储，不存明文，如BCrypt加密） |
| `phone`         | varchar(20)  | 手机号（唯一约束，用于登录/找回密码）        |
| `email`         | varchar(100) | 邮箱（用于登录/通知）                        |
| `avatar`        | varchar(255) | 头像图片路径                                 |
| `status`        | tinyint(1)   | 账号状态（1=正常，0=禁用）                   |
| `create_time`   | datetime     | 注册时间（默认当前时间）                     |
| `update_time`   | datetime     | 信息更新时间（默认当前时间，更新时自动刷新） |


#### 6. `carts`（购物车表）
**核心作用**：记录用户添加到购物车的商品信息，关联用户、商品和规格。  
| 字段名       | 类型   | 说明                                                         |
| ------------ | ------ | ------------------------------------------------------------ |
| `id`         | bigint | 主键（自增）                                                 |
| `product_id` | bigint | 关联商品ID（对应`products.id`），标识购物车中的商品          |
| `num`        | bigint | 商品数量（用户添加的数量）                                   |
| `size_id`    | bigint | 关联规格ID（对应`product_sizes.id`），标识商品的规格（如S码） |
| `user_id`    | bigint | 关联用户ID（对应`user.id`），标识该购物车记录属于哪个用户    |


#### 7. `orders`（订单主表）
**核心作用**：存储订单的整体信息（一个订单对应多个商品，关联订单详情）。  
| 字段名        | 类型         | 说明                                                         |
| ------------- | ------------ | ------------------------------------------------------------ |
| `id`          | bigint       | 主键（自增），唯一标识一个订单                               |
| `order_code`  | varchar(255) | 订单编号（唯一约束），业务上的订单唯一标识（如"20231011123456"） |
| `pay_status`  | int          | 支付状态（如0=未支付、1=已支付、2=退款中）                   |
| `ship_status` | int          | 发货状态（如0=未发货、1=已发货、2=已签收）                   |
| `price`       | double       | 订单总金额（所有订单详情的金额总和）                         |
| `create_at`   | datetime     | 订单创建时间                                                 |
| `update_at`   | datetime     | 订单更新时间（如支付/发货状态变更时刷新）                    |


#### 8. `order_details`（订单详情表）
**核心作用**：记录订单中具体包含的商品信息（订单与商品的多对多关联表）。  
| 字段名            | 类型   | 说明                                                         |
| ----------------- | ------ | ------------------------------------------------------------ |
| `id`              | bigint | 主键（自增）                                                 |
| `product_id`      | bigint | 关联商品ID（对应`products.id`），标识订单中的商品            |
| `product_num`     | bigint | 商品数量（该商品在订单中的购买数量）                         |
| `product_size_id` | bigint | 关联规格ID（对应`product_sizes.id`），标识购买时选择的规格   |
| `product_price`   | double | 商品单价（下单时的价格，固定记录，避免后续商品调价影响订单） |
| `order_id`        | bigint | 关联订单ID（对应`orders.id`），标识该详情属于哪个订单        |


#### 9. `payments`（支付信息表）
**核心作用**：存储订单的支付相关信息（如支付方式、支付凭证等）。  
| 字段名                                                       | 类型         | 说明                                                         |
| ------------------------------------------------------------ | ------------ | ------------------------------------------------------------ |
| `id`                                                         | bigint       | 主键（自增）                                                 |
| `payment_name`                                               | varchar(255) | 支付方式名称（如"微信支付"、"支付宝"）                       |
| `payment_sid`                                                | varchar(255) | 支付平台交易号（第三方支付返回的唯一标识，如微信的商户订单号） |
| `payment_status`                                             | tinyint(1)   | 支付状态（1=支付成功，0=支付失败/未支付）                    |
| `payment_image`                                              | varchar(255) | 支付凭证图片路径（如用户上传的转账截图）                     |
| *注：实际业务中通常会增加`order_id`字段关联`orders.id`，明确支付与订单的对应关系。* |              |                                                              |


#### 10. `seckill_activities`（秒杀活动表）
**核心作用**：管理商品的秒杀活动，包含活动时间、库存、价格等关键信息。  
| 字段名            | 类型         | 说明                                                         |
| ----------------- | ------------ | ------------------------------------------------------------ |
| `id`              | bigint       | 主键（自增），唯一标识一个秒杀活动                           |
| `activity_name`   | varchar(100) | 秒杀活动名称（如"618限时秒杀"）                              |
| `product_id`      | bigint       | 关联商品ID（对应`products.id`，外键约束），指定参与秒杀的商品 |
| `seckill_price`   | double       | 秒杀价格（通常低于商品原价`product_price`）                  |
| `total_stock`     | bigint       | 秒杀总库存（活动可售数量）                                   |
| `remaining_stock` | bigint       | 剩余库存（动态减少，避免超卖）                               |
| `start_time`      | datetime     | 活动开始时间（精确到秒，控制秒杀开启）                       |
| `end_time`        | datetime     | 活动结束时间（控制秒杀关闭）                                 |
| `status`          | tinyint(1)   | 活动状态（0=未开始，1=进行中，2=已结束，3=已取消）           |
| `create_time`     | datetime     | 活动创建时间（默认当前时间）                                 |


### 二、表关系总览
各表通过关联字段形成依赖关系，共同支撑电商业务流程，核心关系如下：


#### 1. 商品相关表（`products`为核心）
- `products` ← `product_sizes`：**一对多**（1个商品有多个规格，如T恤有S/M/L码）。  
  关联字段：`product_sizes.size_product_id` → `products.id`。

- `products` ← `product_images`：**一对多**（1个商品有多个图片，如封面图、细节图）。  
  关联字段：`product_images.image_product_id` → `products.id`。

- `products` ← `product_seos`：**一对一**（1个商品对应1套SEO信息）。  
  关联字段：`product_seos.seo_product_id` → `products.id`。

- `products` ← `seckill_activities`：**一对多**（1个商品可参与多个秒杀活动，如同一商品在不同时间段秒杀）。  
  关联字段：`seckill_activities.product_id` → `products.id`（外键约束）。


#### 2. 用户与购物车（`user`为核心）
- `user` ← `carts`：**一对多**（1个用户有多个购物车记录，对应不同商品/规格）。  
  关联字段：`carts.user_id` → `user.id`。

- `carts` ← `products`：**多对一**（多个购物车记录可关联同一商品）。  
  关联字段：`carts.product_id` → `products.id`。

- `carts` ← `product_sizes`：**多对一**（多个购物车记录可关联同一规格）。  
  关联字段：`carts.size_id` → `product_sizes.id`。


#### 3. 订单相关表（`orders`为核心）
- `orders` ← `order_details`：**一对多**（1个订单包含多个商品，对应多条详情记录）。  
  关联字段：`order_details.order_id` → `orders.id`。

- `order_details` ← `products`：**多对一**（多个订单详情可关联同一商品）。  
  关联字段：`order_details.product_id` → `products.id`。

- `order_details` ← `product_sizes`：**多对一**（多个订单详情可关联同一规格）。  
  关联字段：`order_details.product_size_id` → `product_sizes.id`。

- `orders` ← `payments`：**一对一**（1个订单对应1条支付记录，实际需通过`order_id`关联）。


### 三、业务流程串联
这些表通过关系形成完整的电商流程：  
用户（`user`）浏览商品（`products`+`product_images`+`product_sizes`）→ 添加到购物车（`carts`）→ 下单生成订单（`orders`+`order_details`）→ 支付（`payments`）→ 商家发货；同时，商品可参与秒杀活动（`seckill_activities`），用户秒杀成功后生成特殊订单（价格为秒杀价）。