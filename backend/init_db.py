from app.database import Base, engine
from app.models import AscensionRecord, Record, ResonanceRecord

TABLE_FIELDS = {
    "tacet_records": (
        "id: 主键",
        "date: 日期",
        "player_id: 玩家ID",
        "gold_tubes: 金色密音筒数量",
        "purple_tubes: 紫色密音筒数量",
        "claim_count: 领取次数（1/2）",
        "sola_level: 索拉等级",
        "created_at: 创建时间",
    ),
    "ascension_records": (
        "id: 主键",
        "date: 日期",
        "player_id: 玩家ID",
        "sola_level: 索拉等级",
        "drop_count: 突破材料掉落数量",
        "created_at: 创建时间",
    ),
    "resonance_records": (
        "id: 主键",
        "date: 日期",
        "player_id: 玩家ID",
        "sola_level: 索拉等级",
        "gold: 金色掉落数量",
        "purple: 紫色掉落数量",
        "blue: 蓝色掉落数量",
        "green: 绿色掉落数量",
        "created_at: 创建时间",
    ),
}


def print_table_fields() -> None:
    print("\n表结构：")
    for table_name, fields in TABLE_FIELDS.items():
        print(f"- 表名: {table_name}")
        print("- 字段:")
        for field in fields:
            print(f"  - {field}")
        print()


def init_db() -> None:
    """初始化数据库表并输出表结构。"""
    # Import models so SQLAlchemy registers all tables before create_all.
    _ = (AscensionRecord, Record, ResonanceRecord)

    print("正在创建数据库表...")
    Base.metadata.create_all(bind=engine)
    print("数据库表创建成功！")
    print_table_fields()


if __name__ == "__main__":
    init_db()
