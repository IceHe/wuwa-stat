"""
初始化数据库表
"""
from app.database import engine, Base
from app.models import AscensionRecord, Record, ResonanceRecord

def init_db():
    print("正在创建数据库表...")
    Base.metadata.create_all(bind=engine)
    print("数据库表创建成功！")

    # 显示表结构
    print("\n表结构：")
    print("- 表名: tacet_records")
    print("- 字段:")
    print("  - id: 主键")
    print("  - date: 日期")
    print("  - player_id: 玩家ID")
    print("  - gold_tubes: 金色密音筒数量")
    print("  - purple_tubes: 紫色密音筒数量")
    print("  - sola_level: 索拉等级")
    print("  - created_at: 创建时间")

    print("\n- 表名: ascension_records")
    print("- 字段:")
    print("  - id: 主键")
    print("  - date: 日期")
    print("  - player_id: 玩家ID")
    print("  - sola_level: 索拉等级")
    print("  - drop_count: 突破材料掉落数量")
    print("  - created_at: 创建时间")

    print("\n- 表名: resonance_records")
    print("- 字段:")
    print("  - id: 主键")
    print("  - date: 日期")
    print("  - player_id: 玩家ID")
    print("  - sola_level: 索拉等级")
    print("  - gold: 金色掉落数量")
    print("  - purple: 紫色掉落数量")
    print("  - blue: 蓝色掉落数量")
    print("  - green: 绿色掉落数量")
    print("  - created_at: 创建时间")

if __name__ == "__main__":
    init_db()
