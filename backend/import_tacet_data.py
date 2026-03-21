import re
from datetime import datetime
from sqlalchemy import text
from app.database import SessionLocal
from app.models import Record

# 清空表
db = SessionLocal()
db.execute(text("DELETE FROM tacet_records"))
db.commit()
print("已清空 tacet_records 表")

# 解析导入
count = 0
with open("data/data_20260318.txt", "r", encoding="utf-8") as f:
    next(f)  # 跳过表头
    for line in f:
        line = line.strip()
        if not line:
            continue

        parts = line.split("\t")
        if len(parts) < 3:
            continue

        # 解析日期 2025/8/17
        date_str = parts[0]
        date = datetime.strptime(date_str, "%Y/%m/%d").date()

        # 解析玩家ID
        player_id = parts[1].strip()

        # 解析产出 "索 8 金 3 紫 4"
        output_str = parts[2]
        match = re.search(r"索\s*(\d+)\s*金\s*(\d+)\s*紫\s*(\d+)", output_str)
        if match:
            sola_level = int(match.group(1))
            gold_tubes = int(match.group(2))
            purple_tubes = int(match.group(3))

            record = Record(
                date=date,
                player_id=player_id,
                sola_level=sola_level,
                gold_tubes=gold_tubes,
                purple_tubes=purple_tubes
            )
            db.add(record)
            count += 1

            # 每1000条提交一次
            if count % 1000 == 0:
                db.commit()
                print(f"已导入 {count} 条...")

db.commit()
print(f"共导入 {count} 条数据")
db.close()
