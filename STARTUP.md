# 启动指南

## 当前状态

✅ **后端已启动并运行正常**
- 地址：http://localhost:8000
- API文档：http://localhost:8000/docs
- 数据库已创建并导入示例数据

## 启动前端

由于网络问题，前端依赖安装失败。请手动执行以下步骤：

### 方法1：使用淘宝镜像（推荐）

```bash
cd /Users/icehe/Projects/claude-1st-try/frontend

# 清理旧的依赖
rm -rf node_modules package-lock.json

# 使用淘宝镜像安装
npm install --registry=https://registry.npmmirror.com

# 启动开发服务器
npm run dev
```

### 方法2：配置npm镜像后安装

```bash
cd /Users/icehe/Projects/claude-1st-try/frontend

# 设置淘宝镜像
npm config set registry https://registry.npmmirror.com

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 方法3：使用cnpm

```bash
# 安装cnpm
npm install -g cnpm --registry=https://registry.npmmirror.com

cd /Users/icehe/Projects/claude-1st-try/frontend

# 使用cnpm安装
cnpm install

# 启动开发服务器
npm run dev
```

## 访问应用

前端启动成功后，访问：http://localhost:5173

## 功能说明

### 1. 录入数据
- 选择日期
- 输入玩家ID
- 输入金色/紫色密音筒数量
- 选择索拉等级
- 设置录入次数（支持批量录入相同数据）

### 2. 查看记录
- 查看所有产出记录
- 按玩家ID筛选
- 按索拉等级筛选

### 3. 统计数据
- 总记录数
- 玩家数量
- 金色/紫色密音筒总数和平均值

## 测试API

后端API已经可以直接使用：

```bash
# 获取所有记录
curl --noproxy localhost http://localhost:8000/api/records

# 获取统计数据
curl --noproxy localhost http://localhost:8000/api/stats

# 批量创建记录
curl --noproxy localhost -X POST http://localhost:8000/api/records \
  -H "Content-Type: application/json" \
  -d '{
    "records": [
      {
        "date": "2025-08-18",
        "player_id": "120003177",
        "gold_tubes": 5,
        "purple_tubes": 3,
        "sola_level": 8
      }
    ]
  }'
```

## 停止服务

### 停止后端
后端正在后台运行，如需停止：
```bash
# 查找进程
ps aux | grep uvicorn

# 停止进程（替换PID为实际进程号）
kill <PID>
```

### 停止前端
在前端运行的终端按 `Ctrl+C`

## 数据库管理

### 查看数据
```bash
psql -U icehe -d postgres -c "SELECT * FROM records;"
```

### 清空数据
```bash
psql -U icehe -d postgres -c "TRUNCATE TABLE records;"
```

## 项目结构

```
.
├── backend/              # FastAPI后端
│   ├── app/
│   │   ├── main.py      # 主应用
│   │   ├── models.py    # 数据库模型
│   │   ├── schemas.py   # API schemas
│   │   ├── database.py  # 数据库配置
│   │   └── api/
│   │       └── routes.py # API路由
│   ├── init_db.py       # 初始化数据库
│   ├── import_sample_data.py # 导入示例数据
│   └── .env             # 环境配置
└── frontend/            # Vue 3前端
    ├── src/
    │   ├── App.vue
    │   ├── main.ts
    │   ├── components/  # 组件
    │   └── api/         # API调用
    └── vite.config.ts
```
