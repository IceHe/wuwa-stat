# 启动指南

适用于“鸣潮统计网页工具”，当前包含：
- 无音区产出统计
- 共鸣者突破材料统计
- 凝素领域产出统计

## 启动后端

```bash
cd /root/wuwa/wuwa_stat/backend

cp .env.example .env
# 按实际情况编辑 .env 中的 DATABASE_URL 和 FRONTEND_URL

pip install -r requirements.txt
python init_db.py
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

启动后访问：
- 应用接口：`http://localhost:8000`
- API 文档：`http://localhost:8000/docs`
- 健康检查：`http://localhost:8000/health`

## 启动前端

```bash
cd /root/wuwa/wuwa_stat/frontend
npm install
npm run dev
```

前端默认地址：`http://localhost:5173`

如果 npm 网络较慢，可改用镜像：

```bash
cd /root/wuwa/wuwa_stat/frontend
npm install --registry=https://registry.npmmirror.com
npm run dev
```

## 鉴权说明

项目默认接入外部鉴权服务，接口会校验 Token 权限。

- 默认鉴权服务地址：`http://127.0.0.1:8080`
- `view` 权限：允许查看列表和统计
- `edit` 权限：允许新增和删除记录
- `manage` 权限：包含全部权限

如果鉴权服务不可用，前端登录和后端受保护接口都会失败。

## 页面功能

### 无音区产出统计

- 录入金色/紫色密音筒产出
- 支持单次和双次领取
- 支持列表筛选、分页和统计分析

### 共鸣者突破材料统计

- 录入每次突破材料掉落数量
- 支持列表筛选、分页和按索拉等级统计

### 凝素领域产出统计

- 录入金、紫、蓝、绿四档掉落数量
- 支持列表筛选、分页和组合分布统计

## 常用接口示例

以下示例假设你已经准备好可用 Token，并替换 `${TOKEN}`。

```bash
curl --noproxy localhost \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8000/api/tacet_records"

curl --noproxy localhost \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8000/api/ascension-records"

curl --noproxy localhost \
  -H "Authorization: Bearer ${TOKEN}" \
  "http://localhost:8000/api/resonance-detailed-stats"

curl --noproxy localhost -X POST "http://localhost:8000/api/tacet_records" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "tacet_records": [
      {
        "date": "2025-08-18",
        "player_id": "120003177",
        "gold_tubes": 7,
        "purple_tubes": 8,
        "claim_count": 2,
        "sola_level": 8
      }
    ]
  }'
```

## 停止服务

- 停止后端：结束运行 `uvicorn` 的终端或进程
- 停止前端：结束运行 `npm run dev` 的终端

## 数据库备注

当前会自动创建以下表：
- `tacet_records`
- `ascension_records`
- `resonance_records`

如果你是从旧版无音区表升级：

```bash
psql -U <user> -d <db> -c "ALTER TABLE tacet_stats RENAME TO tacet_records;"
psql -U <user> -d <db> -c "ALTER INDEX IF EXISTS idx_tacet_stats_date RENAME TO idx_tacet_records_date;"
psql -U <user> -d <db> -c "ALTER INDEX IF EXISTS idx_tacet_stats_player_id RENAME TO idx_tacet_records_player_id;"
psql -U <user> -d <db> -c "ALTER TABLE tacet_records ADD COLUMN IF NOT EXISTS claim_count INTEGER NOT NULL DEFAULT 1;"
```
