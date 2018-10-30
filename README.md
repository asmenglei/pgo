# PGO
PGO应用框架即"Pinguo GO application framework"，是Camera360广告服务端团队研发的一款简单、高性能、组件化的GO应用框架。受益于GO语言高性能与原生协程，业务从php+yii2升级到PGO后，线上表现单机处理能力提高5-10倍，从实际使用中看其开发效率亦不输于PHP。


参考文档：[pgo-docs](https://github.com/pinguo/pgo-docs)

应用示例：[pgo-demo](https://github.com/pinguo/pgo-demo)

## 环境要求
- GO 1.10+
- Make 3.8+
- Linux/MacOS
- Glide 0.13+ (建议)
- GoLand 2018 (建议)

## 项目目录
规范：
- 一个项目为一个独立的目录，不使用GO全局工作空间。
- 项目的GOPATH为项目根目录，不要依赖系统的GOPATH。
- 除GO标准库外，所有外部依赖代码放到"src/vendor"下。
- 项目源码文件与目录使用大写驼峰(CamelCase)形式。

```
<project>
├── bin/                # 编译程序目录
├── conf/               # 配置文件目录
│   ├── production/     # 环境配置目录
│   │   ├── app.json
│   │   └── params.json
│   ├── testing/
│   ├── app.json        # 项目配置文件
│   └── params.json     # 自定义配置文件
├── makefile            # 编译打包
├── runtime/            # 运行时目录
├── public/             # 静态资源目录
├── view/               # 视图模板目录
└── src/                # 项目源码目录
    ├── Command/        # 命令行控制器目录
    ├── Controller/     # HTTP控制器目录
    ├── Lib/            # 项目基础库目录
    ├── Main/           # 项目入口目录
    ├── Model/          # 模型目录(数据交互)
    ├── Service/        # 服务目录(业务逻辑)
    ├── Struct/         # 结构目录(数据定义)
    ├── Test/           # 测试目录(单测/性能)
    ├── vendor/         # 第三方依赖目录
    ├── glide.lock      # 项目依赖锁文件
    └── glide.yaml      # 项目依赖配置文件
```

## 依赖管理
建议使用glide做为依赖管理工具(类似php的composer)，不使用go官方的dep工具

安装(mac)：`brew install glide`

使用(调用目录为项目的src目录)：
```
glide init              # 初始化项目
glide get <pkg>         # 下载pkg并添加依赖
    --all-dependencies  # 下载pkg的所有依赖
glide get <pkg>#v1.2    # 下载指定版本的pkg
glide install           # 根据lock文件下载依赖
glide update            # 更新依赖包
```

## 基准测试
TODO

## 快速开始
1. 创建项目目录(以下三种方法均可)
    - 参见《项目目录》手动创建
    - 从[pgo-demo](https://github.com/pinguo/pgo-demo)克隆目录结构
    - 拷贝makefile至项目根目录，执行`make init`创建目录
2. 修改配置文件(conf/app.json)
    ```json
    {
        "name": "pgo-demo",
        "GOMAXPROCS": 2,
        "runtimePath": "@app/runtime",
        "publicPath": "@app/public",
        "viewPath": "@app/view",
        "server": {
            "addr": "0.0.0.0:8000",
            "readTimeout": "30s",
            "writeTimeout": "30s",
            "plugins": []
        },
        "components": {
            "log": {
                "levels": "ALL",
                "targets": {
                    "info": {
                        "class": "@pgo/FileTarget",
                        "levels": "DEBUG,INFO,NOTICE",
                        "filePath": "@runtime/info.log",
                    },
                    "error": {
                        "class": "@pgo/FileTarget",
                        "levels": "WARN,ERROR,FATAL",
                        "filePath": "@runtime/error.log",
                    },
                    "console": {
                        "class": "@pgo/ConsoleTarget",
                        "levels": "ALL"
                    }
                }
            }
        }
    }
    ```
3. 安装PGO(以下两种方法均可)
    - 在项目根目录执行`cd src && glide init && glide get github.com/pinguo/pgo`
    - 如果已拷贝makefile，在项目根目录执行`make pgo`
4. 创建控制器(src/Controller/WelcomeController.go)
    ```go
    package Controller

    import (
        "github.com/pinguo/pgo"
        "net/http"
        "time"
    )

    type WelcomeController struct {
        pgo.Controller
    }

    func (w *WelcomeController) ActionIndex() {
        data := pgo.Map{"text": "welcome to pgo-demo", "now": time.Now()}
        w.OutputJson(data, http.StatusOK)
    }
    ```
5. 注册控制器(src/Controller/Init.go)
    ```go
    package Controller

    import "github.com/pinguo/pgo"

    func init() {
        container := pgo.App.GetContainer()

        container.Bind(&WelcomeController{})
    }
    ```
6. 创建程序入口(src/Main/main.go)
    ```go
    package main

    import (
        _ "Controller" // 导入控制器

        "github.com/pinguo/pgo"
    )

    func main() {
        pgo.Run() // 运行程序
    }
    ```
7. 编译运行
    ```sh
    make start
    curl http://127.0.0.1:8000/welcome
    ```

## 使用示例
### 项目配置
- 项目配置文件`conf/app.json`, 可任意添加自定义配置文件如params.json
- 目前仅支持json配置文件，后续会支持yaml配置文件
- 所有配置文件均是一个json对象
- 支持任意环境目录，环境目录中的同名字段会覆盖基础配置中的字段
- 通过bin/<binName> --env production指定程序运行环境
- 配置都有默认值，配置文件中的值会覆盖默认值(默认值参见组件说明)
- 配置文件支持环境变量，格式`${envName|default}`，当envName不存在时使用default
- 配置文件中路径及类名支持另名字符串，PGO定义的别名如下：
    - `@app` 项目根目录绝对路径
    - `@runtime` 项目运行时目录绝对路径
    - `@view` 项目视图模板目录绝对路径
    - `@pgo` PGO框架import路径

使用配置：
```go
cfg := pgo.App.GetConfig() // 获取配置对象
name := cfg.GetString("app.name", "demo") // 获取String，不存在返回"demo"
procs := cfg.GetInt("app.GOMAXPROCS", 2) // 获取Integer, 不存在返回2
price := cfg.GetFloat("params.goods.basePrice", 0) // 获取Float, 不存在返回0
enable := cfg.GetBool("params.detect.enable", false) // 获取Bool, 不存在返回false

// 除基本类型外，通过Get方法获取原始配置数据，需要进行类型转换
plugins, ok := cfg.Get("app.servers.plugins").([]interface{}) // 获取数组
log, ok := cfg.Get("app.conponents.log").(map[string]interface{}) // 获取对象
```

