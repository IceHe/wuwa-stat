"""
导入示例数据
"""
from datetime import datetime
from app.database import SessionLocal
from app.models import Record

def import_sample_data():
    db = SessionLocal()

    # 示例数据
    sample_data = [
        {"date": "2025-08-17", "player_id": "120003177", "sola_level": 8, "gold_tubes": 3, "purple_tubes": 4},
        {"date": "2025-08-17", "player_id": "120003177", "sola_level": 8, "gold_tubes": 3, "purple_tubes": 4},
        {"date": "2025-08-17", "player_id": "122075512", "sola_level": 8, "gold_tubes": 3, "purple_tubes": 4},
        {"date": "2025-08-17", "player_id": "122075512", "sola_level": 8, "gold_tubes": 4, "purple_tubes": 4},
        {"date": "2025-08-17", "player_id": "108119803", "sola_level": 8, "gold_tubes": 3, "purple_tubes": 4},
    ]

    try:
        records = []
        for data in sample_data:
            record = Record(**data)
            records.append(record)

        db.add_all(records)
        db.commit()

        print(f"成功导入 {len(records)} 条示例数据！")

        # 显示导入的数据
        print("\n导入的数据：")
        for record in records:
            print(f"  日期: {record.date}, 玩家ID: {record.player_id}, "
                  f"索拉等级: {record.sola_level}, "
                  f"金色: {record.gold_tubes}, 紫色: {record.purple_tubes}")

    except Exception as e:
        db.rollback()
        print(f"导入失败: {e}")
    finally:
        db.close()

if __name__ == "__main__":
    import_sample_data()
