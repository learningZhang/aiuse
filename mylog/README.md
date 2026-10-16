# mylog —— 用户日记系统（Elasticsearch 学习项目）

Go + Gin + Elasticsearch 7.8 的练手项目：把每天的所见所闻写成日记存入 ES，支持标签分类、全文检索、分页浏览。

## 运行

```bash
make es-init     # 创建 diary 索引
make run         # 启动服务，监听 :8080
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/tags
```

## 目录结构（对齐 userhub 分层）

```
main.go              入口：配置 → 日志 → ES → 路由 → 启动
conf/                配置：yaml + 环境变量覆盖
provider/            基础设施：日志、ES 客户端、RequestID 上下文
router/              路由注册 + 中间件（RequestID/Logger/Recovery/CORS）
app/controller/      控制器：绑定参数 → 校验 → 调 service → 错误映射 → 响应
app/service/         业务逻辑：参数归一化、默认值填充、流程编排
app/repository/      ES 数据访问：一索引一文件，只构造 DSL 与解析响应
app/model/           请求/响应/文档结构体、业务常量
app/exception/       自定义异常（值类型，支持 type-switch）
utils/               纯工具函数，不依赖业务层
scripts/             ES 索引 mapping 与测试数据
```

## 依赖方向（严格单向）

```
main.go → router → app/controller → app/service → app/repository → provider
                        ↘ app/model / utils / conf / provider 任意层可用
```

## 配置

主配置 `conf/config/config.yaml`。敏感项用环境变量注入，不入库：

| 环境变量 | 说明 |
|---|---|
| `MYLOG_SERVER_PORT` | 监听端口 |
| `MYLOG_SERVER_MODE` | debug / release / test |
| `MYLOG_ES_ADDRESSES` | ES 地址（单个） |
| `MYLOG_ES_USERNAME` | ES 用户名 |
| `MYLOG_ES_PASSWORD` | ES 密码 |

## 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/health` | 服务与 ES 连通性检查 |
| GET | `/tags` | terms 聚合：全部标签及篇数 |
| GET | `/diaries?tag=&page=&size=` | 列表：按 date 倒序 + from/size 分页，可选标签过滤 |
| GET | `/diaries/:id` | 详情：按文档 ID 取完整字段 |
| GET | `/search?keyword=&tag=&page=&size=` | 全文检索：multi_match(summary^2, content)，可选标签叠加 |

> 注：搜索接口为 `/search` 而非 `/diaries/search`——Gin 不允许同层级静态段与 `:id` 通配共存（注册时 panic）。

## 响应约定

成功：直接返回需求文档定义的业务结构 + HTTP 200。
失败：

```json
{
  "error": {
    "code": "InvalidParameter",
    "message": "参数错误说明",
    "request_id": "uuid"
  }
}
```

错误码：`InvalidParameter`(400) / `NotFound`(404) / `EsUnavailable`(502) / `InternalError`(500)。
