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
