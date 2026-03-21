# 鸣潮无音区产出统计工具

统计鸣潮游戏无音区的金色和紫色密音筒产出数据。

## 技术栈

- **后端**: FastAPI + SQLAlchemy + PostgreSQL
- **前端**: Vue 3 + TypeScript + Element Plus + Vite

## 项目结构

```
├── backend/          # FastAPI 后端
│   ├── app/
│   │   ├── main.py       # 主应用
│   │   ├── models.py     # 数据库模型
│   │   ├── schemas.py    # Pydantic schemas
│   │   ├── database.py   # 数据库连接
│   │   └── api/
│   │       └── routes.py # API 路由
│   ├── requirements.txt
│   └── .env.example
└── frontend/         # Vue 3 前端
    ├── src/
    │   ├── main.ts
    │   ├── App.vue
    │   ├── components/   # 组件
    │   └── api/          # API 调用
    └── package.json
```

## 快速开始

### 1. 数据库配置

创建 PostgreSQL 数据库：

```bash
createdb wuthering_waves
```

### 2. 后端设置

```bash
cd backend

# 复制环境配置文件
cp .env.example .env

# 编辑 .env 文件，填入你的数据库连接信息
# DATABASE_URL=postgresql://username:password@localhost:5432/wuthering_waves

# 安装依赖
pip install -r requirements.txt

# 启动后端服务
uvicorn app.main:app --reload --port 8001
```

后端将运行在 http://localhost:8001

### 3. 前端设置

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将运行在 http://localhost:5173

## 功能特性

### 数据录入
- 支持批量录入相同数据（例如双倍领取）
- 记录字段：
  - 日期
  - 玩家ID
  - 金色密音筒数量
  - 紫色密音筒数量
  - 索拉等级（默认8级）

### 数据查询
- 按玩家ID筛选
- 按索拉等级筛选
- 按日期范围筛选
- 分页显示

### 统计分析
- 总记录数
- 玩家数量
- 金色/紫色密音筒总数
- 平均产出数量

## API 文档

启动后端后访问：http://localhost:8000/docs

## 数据库表结构

如果你是从旧表 `tacet_stats` 升级，请先执行：

```sql
ALTER TABLE tacet_stats RENAME TO tacet_records;
```

如需保持索引命名一致性，可按实际索引名再执行重命名。

```sql
ALTER INDEX IF EXISTS idx_tacet_stats_date RENAME TO idx_tacet_records_date;
ALTER INDEX IF EXISTS idx_tacet_stats_player_id RENAME TO idx_tacet_records_player_id;
```

当前表结构：

```sql
CREATE TABLE tacet_records (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    player_id VARCHAR NOT NULL,
    gold_tubes INTEGER NOT NULL DEFAULT 0,
    purple_tubes INTEGER NOT NULL DEFAULT 0,
    sola_level INTEGER NOT NULL DEFAULT 8,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tacet_records_date ON tacet_records(date);
CREATE INDEX idx_tacet_records_player_id ON tacet_records(player_id);
```

## 开发说明

### 后端开发
- 使用 FastAPI 的自动文档功能测试 API
- 数据库迁移使用 SQLAlchemy 的 `Base.metadata.create_all()`

### 前端开发
- 使用 Element Plus 组件库
- API 请求通过 Vite 代理转发到后端
- TypeScript 提供类型安全

## 部署

### 后端部署
```bash
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

### 前端部署
```bash
npm run build
# 将 dist/ 目录部署到静态服务器
```

## 许可证

MIT
