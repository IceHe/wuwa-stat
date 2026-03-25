# 鸣潮统计网页工具

用于录入、查询和统计《鸣潮》多类副本产出数据的网页工具，目前包含：
- 无音区产出统计
- 共鸣者突破材料统计
- 凝素领域产出统计

## 技术栈

- **后端**: FastAPI + SQLAlchemy + PostgreSQL
- **前端**: Vue 3 + TypeScript + Element Plus + Vite
- **鉴权**: 基于外部鉴权服务的 Token 权限校验

## 项目结构

```text
├── backend/              # FastAPI 后端
│   ├── app/
│   │   ├── main.py       # 主应用与 OpenAPI 配置
│   │   ├── models.py     # 数据库模型
│   │   ├── schemas.py    # Pydantic schemas
│   │   ├── database.py   # 数据库连接与配置
│   │   ├── auth.py       # Token 鉴权
│   │   └── api/
│   │       └── routes.py # 三类统计模块 API
│   ├── data/             # 导入脚本使用的数据文件
│   ├── init_db.py        # 初始化数据库
│   ├── import_*.py       # 数据导入脚本
│   ├── requirements.txt
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

pip install -r requirements.txt
python init_db.py
uvicorn app.main:app --reload --port 8000
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

## API 文档

启动后访问：`http://localhost:8000/docs`

## 数据模型

- `tacet_records`：无音区产出记录
- `ascension_records`：共鸣者突破材料记录
- `resonance_records`：凝素领域产出记录

如果你是从旧版无音区表 `tacet_stats` 升级，请先执行：

```sql
ALTER TABLE tacet_stats RENAME TO tacet_records;
ALTER INDEX IF EXISTS idx_tacet_stats_date RENAME TO idx_tacet_records_date;
ALTER INDEX IF EXISTS idx_tacet_stats_player_id RENAME TO idx_tacet_records_player_id;
ALTER TABLE tacet_records ADD COLUMN IF NOT EXISTS claim_count INTEGER NOT NULL DEFAULT 1;
```

## 开发说明

### 后端开发

- API 位于 `backend/app/api/routes.py`
- 数据库模型位于 `backend/app/models.py`
- 当前通过 `Base.metadata.create_all()` 建表，未引入独立迁移框架
- OpenAPI 文档标题与描述由 `backend/app/main.py` 提供

### 前端开发

- 入口页在 `frontend/src/App.vue`
- 三类统计模块分别对应各自的 Input/List/Stats 组件
- API 请求通过 Vite 代理到 `http://localhost:8000`
- TypeScript 提供类型安全

## 部署

### 后端部署

```bash
cd backend
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

### 前端部署

```bash
cd frontend
npm install
npm run build
```

将 `dist/` 目录部署到静态文件服务器即可。

## 许可证

MIT
