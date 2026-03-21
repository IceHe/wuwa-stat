import re
from datetime import datetime
from pathlib import Path

from app.database import SessionLocal
from app.models import ResonanceRecord

TRANSCRIPT_PATH = Path("/root/.claude/projects/-root-wuwa-wuwa-stat/fd3333aa-9a41-4956-835a-021324f12ffa.jsonl")
BATCH_SIZE = 1000


def parse_date(date_text: str):
    return datetime.strptime(date_text, "%Y年%m月%d日").date()


def extract_dataset_text() -> str:
    import json

    needle = "新增统计tab:凝素领域产出统计"
    with TRANSCRIPT_PATH.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                obj = json.loads(line)
            except Exception:
                continue
            message = obj.get("message")
            if not isinstance(message, dict):
                continue
            content = message.get("content")
            if isinstance(content, str) and content.startswith(needle):
                return content

    raise RuntimeError("未在 transcript 中找到凝素领域数据")


def parse_records(dataset_text: str):
    records = []
    failed = []

    pattern = re.compile(
        r"^\s*(\d+)\s+(\d{4}年\d{1,2}月\d{1,2}日)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s*$"
    )

    for line in dataset_text.splitlines():
        line = line.strip()
        if not line:
            continue
        if line.startswith("新增统计tab:") or line.startswith("索拉等级"):
            continue

        match = pattern.match(line)
        if not match:
            failed.append(line)
            continue

        sola_level, date_text, gold, purple, blue, green, player_id = match.groups()

        records.append(
            ResonanceRecord(
                sola_level=int(sola_level),
                date=parse_date(date_text),
                gold=int(gold),
                purple=int(purple),
                blue=int(blue),
                green=int(green),
                player_id=player_id,
            )
        )

    return records, failed


def main():
    dataset_text = extract_dataset_text()
    records, failed = parse_records(dataset_text)

    if not records:
        raise RuntimeError("没有解析到任何有效记录")

    db = SessionLocal()
    try:
        db.query(ResonanceRecord).delete()
        db.commit()

        inserted = 0
        for record in records:
            db.add(record)
            inserted += 1
            if inserted % BATCH_SIZE == 0:
                db.commit()
                print(f"已导入 {inserted} 条...")

        db.commit()
        print(f"导入完成，共 {inserted} 条")
        print(f"无法解析行数: {len(failed)}")

        if failed:
            print("示例异常行:")
            for sample in failed[:5]:
                print(sample)
    finally:
        db.close()


if __name__ == "__main__":
    main()
