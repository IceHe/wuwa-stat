# 鸣潮统计网页工具

用于录入、查询和统计《鸣潮》多类副本产出数据的网页工具，目前包含：
- 无音区产出统计
- 共鸣者突破材料统计
- 凝素领域产出统计

## 技术栈

- **后端**: Go + net/http + database/sql + PostgreSQL
- **前端**: Vue 3 + TypeScript + Element Plus + Vite
- **鉴权**: 基于外部鉴权服务的 Token 权限校验

## 项目结构

```text
├── backend/              # Go 后端
│   ├── cmd/
│   │   ├── server/       # HTTP 服务入口
│   │   └── initdb/       # 初始化数据库入口
│   ├── internal/
│   │   ├── api/          # 路由、鉴权、响应结构
│   │   ├── config/       # .env 与配置加载
│   │   └── db/           # 数据库连接与建表/补字段
│   ├── go.mod
│   └── .env.example
└── frontend/             # Vue 3 前端
    ├── src/
    │   ├── main.ts
    │   ├── App.vue       # 三个统计标签页入口
    │   ├── components/   # 录入、列表、统计组件
    │   └── api/          # API 调用与鉴权状态
	└── package.json
```

说明：项目只有顶层 `frontend/` 这一个前端工程。

## 快速开始

需要 `Go 1.22+`。

### 1. 数据库配置

创建 PostgreSQL 数据库：

```bash
createdb wuthering_waves
```

### 2. 后端设置

```bash
cd backend

cp .env.example .env
# 编辑 .env 中的 DATABASE_URL 和 FRONTEND_URL

go run ./cmd/initdb
go run ./cmd/server
```

后端将运行在 `http://localhost:8000`

### 3. 前端设置

```bash
cd frontend
npm install
npm run dev
```

前端将运行在 `http://localhost:5173`

如果 npm 访问较慢，可以使用镜像：

```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
npm run dev
```

### 4. 鉴权说明

前后端接口默认需要 Token 鉴权，后端会向外部鉴权服务校验权限。

- 需要至少 `view` 权限才能查看数据
- 需要 `edit` 或 `manage` 权限才能新增、删除记录
- 默认鉴权服务地址为 `http://127.0.0.1:8080`

## 功能特性

### 无音区产出统计

- 记录金色/紫色密音筒产出
- 支持领取 1 次和领取 2 次录入
- 支持按玩家 ID、索拉等级、日期范围筛选
- 提供基础统计和按索拉等级拆分的详细统计

### 共鸣者突破材料统计

- 记录每次掉落的突破材料数量
- 支持按玩家 ID、索拉等级、日期范围筛选
- 提供按索拉等级分组的掉落分布和均值统计

### 凝素领域产出统计

- 记录金、紫、蓝、绿四档掉落数量
- 支持按玩家 ID、索拉等级、日期范围筛选
- 提供按索拉等级分组的组合分布和均值统计

### 通用能力

- Token 登录与权限控制
- 分页列表查询
- 删除单条记录
- API 通过 Vite 代理到后端

## 数据模型

- `tacet_records`：无音区产出记录
- `ascension_records`：共鸣者突破材料记录
- `resonance_records`：凝素领域产出记录

## 开发说明

### 后端开发

- HTTP 入口位于 `backend/cmd/server/main.go`
- 接口实现位于 `backend/internal/api/handlers.go`
- 数据库初始化与补字段位于 `backend/internal/db/postgres.go`
- 配置加载位于 `backend/internal/config/config.go`

### 前端开发

- 入口页在 `frontend/src/App.vue`
- 三类统计模块分别对应各自的 Input/List/Stats 组件
- API 请求通过 Vite 代理到 `http://localhost:8000`
- TypeScript 提供类型安全

## 联调自检

启动后端、前端和鉴权服务后，建议至少执行以下检查：

```bash
curl --noproxy '*' http://localhost:8000/health
curl --noproxy '*' -H "Authorization: Bearer ${TOKEN}" http://localhost:8000/api/auth/me
curl --noproxy '*' -H "Authorization: Bearer ${TOKEN}" "http://localhost:8000/api/tacet_records?limit=1"
curl --noproxy '*' -H "Authorization: Bearer ${TOKEN}" "http://localhost:8000/api/ascension-records?limit=1"
curl --noproxy '*' -H "Authorization: Bearer ${TOKEN}" "http://localhost:8000/api/resonance-records?limit=1"
```

预期结果：
- `health` 返回 `{"status":"ok"}`
- 业务接口返回 `200`，且未出现 `Token 无效或已过期` / `鉴权服务不可用`

## 部署

### 后端部署

```bash
cd /root/wuwa/stat/backend
go build -o server ./cmd/server
```

### 前端部署

```bash
cd frontend
npm install
npm run build
```

将 `dist/` 目录部署到静态文件服务器即可。

### 域名部署示例

- 前端域名：`https://stat.icehe.life`
- 推荐在前端域名下额外转发 `/api` 到后端，这样前端可继续使用相对路径请求接口
- 后端 `.env` 中的 `FRONTEND_URL` 需要设置为 `https://stat.icehe.life`
- 生产环境推荐使用 `npm run build` 生成 `dist/`，再由 nginx 直接托管静态文件，而不是长期运行 Vite 开发服务
- systemd 后端服务建议直接执行 Go 二进制：`/root/wuwa/stat/backend/server`
- 可直接复用仓库内模板：[deploy/systemd/wuwa-stat-backend.service](/root/wuwa/stat/deploy/systemd/wuwa-stat-backend.service)

## 许可证

MIT，见 [LICENSE](/root/wuwa/stat/LICENSE)。
