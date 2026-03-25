from datetime import date
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy import func
from sqlalchemy.orm import Session

from app.auth import require_edit_permission, require_view_permission
from app.database import get_db
from app.models import AscensionRecord, Record, ResonanceRecord
from app.schemas import (
    AscensionDetailedStatsResponse,
    AscensionDropCombination,
    AscensionRecordBatchCreate,
    AscensionRecordResponse,
    AscensionSolaLevelStats,
    DetailedStatsResponse,
    DropCombination,
    RecordBatchCreate,
    RecordsListResponse,
    RecordResponse,
    ResonanceDetailedStatsResponse,
    ResonanceDropCombination,
    ResonanceRecordBatchCreate,
    ResonanceRecordResponse,
    ResonanceSolaLevelStats,
    SolaLevelStats,
    StatsResponse,
)

router = APIRouter(prefix="/api")


TACET_SINGLE_COMBOS = {
    8: [(4, 4), (3, 4)],
    7: [(4, 4), (4, 3), (3, 4), (3, 3)],
    6: [(4, 4), (4, 3), (3, 4), (3, 3)],
    5: [(3, 6), (3, 5), (2, 6), (2, 5)],
}


def split_tacet_combination(sola_level: int, gold_tubes: int, purple_tubes: int, claim_count: int) -> list[tuple[int, int]]:
    if claim_count <= 1:
        return [(gold_tubes, purple_tubes)]

    combos = TACET_SINGLE_COMBOS.get(sola_level, [])
    matching_pairs: list[list[tuple[int, int]]] = []

    for left_combo in combos:
        for right_combo in combos:
            if left_combo[0] + right_combo[0] == gold_tubes and left_combo[1] + right_combo[1] == purple_tubes:
                ordered_pair = sorted([left_combo, right_combo], reverse=True)
                matching_pairs.append(ordered_pair)

    if not matching_pairs:
        return [(gold_tubes, purple_tubes)]

    matching_pairs.sort(reverse=True)
    return matching_pairs[0]


@router.post("/tacet_records", response_model=list[RecordResponse], tags=["tacet"])
def create_records(
    batch: RecordBatchCreate,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    """批量创建记录"""
    db_records = []
    for record_data in batch.tacet_records:
        db_record = Record(**record_data.model_dump())
        db_records.append(db_record)

    db.add_all(db_records)
    db.commit()

    for record in db_records:
        db.refresh(record)

    return db_records


@router.get("/tacet_records", response_model=RecordsListResponse, tags=["tacet"])
def get_records(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    sola_level: Optional[int] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=1000),
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    """查询记录，支持筛选和分页"""
    query = db.query(Record)

    if player_id:
        query = query.filter(Record.player_id == player_id)
    if start_date:
        query = query.filter(Record.date >= start_date)
    if end_date:
        query = query.filter(Record.date <= end_date)
    if sola_level:
        query = query.filter(Record.sola_level == sola_level)

    total = query.count()
    records = query.order_by(Record.created_at.desc(), Record.id.desc()).offset(skip).limit(limit).all()

    return {
        "data": records,
        "total": total,
        "page_size": limit,
        "current_page": skip // limit + 1,
    }


@router.get("/stats", response_model=StatsResponse, tags=["tacet"])
def get_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    """获取统计数据"""
    query = db.query(Record)

    if player_id:
        query = query.filter(Record.player_id == player_id)
    if start_date:
        query = query.filter(Record.date >= start_date)
    if end_date:
        query = query.filter(Record.date <= end_date)

    total_records = query.count()

    if total_records == 0:
        return StatsResponse(
            total_records=0,
            total_claim_count=0,
            total_gold_tubes=0,
            total_purple_tubes=0,
            avg_gold_tubes=0.0,
            avg_purple_tubes=0.0,
            player_count=0,
        )

    stats = query.with_entities(
        func.sum(Record.claim_count).label("total_claim_count"),
        func.sum(Record.gold_tubes).label("total_gold"),
        func.sum(Record.purple_tubes).label("total_purple"),
        func.count(func.distinct(Record.player_id)).label("player_count"),
    ).first()

    total_claim_count = int(stats.total_claim_count or 0)
    avg_gold = float(stats.total_gold or 0) / total_claim_count if total_claim_count > 0 else 0.0
    avg_purple = float(stats.total_purple or 0) / total_claim_count if total_claim_count > 0 else 0.0

    return StatsResponse(
        total_records=total_records,
        total_claim_count=total_claim_count,
        total_gold_tubes=stats.total_gold or 0,
        total_purple_tubes=stats.total_purple or 0,
        avg_gold_tubes=avg_gold,
        avg_purple_tubes=avg_purple,
        player_count=stats.player_count or 0,
    )


@router.get("/detailed-stats", response_model=DetailedStatsResponse, tags=["tacet"])
def get_detailed_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    """获取详细统计数据，按索拉等级和密音筒组合分组"""
    query = db.query(Record)

    if player_id:
        query = query.filter(Record.player_id == player_id)
    if start_date:
        query = query.filter(Record.date >= start_date)
    if end_date:
        query = query.filter(Record.date <= end_date)

    grouped_stats = query.with_entities(
        Record.sola_level,
        Record.claim_count,
        Record.gold_tubes,
        Record.purple_tubes,
        func.count().label("count"),
    ).group_by(Record.sola_level, Record.claim_count, Record.gold_tubes, Record.purple_tubes).all()

    level_data: dict[int, dict[tuple[int, int], int]] = {}
    for stat in grouped_stats:
        level = stat.sola_level
        if level not in level_data:
            level_data[level] = {}

        split_combos = split_tacet_combination(
            stat.sola_level,
            stat.gold_tubes,
            stat.purple_tubes,
            stat.claim_count,
        )
        for split_gold, split_purple in split_combos:
            combo_key = (split_gold, split_purple)
            level_data[level][combo_key] = level_data[level].get(combo_key, 0) + stat.count

    level_stats = []
    for level in sorted(level_data.keys(), reverse=True):
        combinations_data = level_data[level]
        total_count = sum(combinations_data.values())
        total_exp = sum((gold * 5000 + purple * 2000) * count for (gold, purple), count in combinations_data.items())
        avg_exp = total_exp / total_count if total_count > 0 else 0

        combinations = []
        for (gold_tubes, purple_tubes), count in sorted(
            combinations_data.items(),
            key=lambda item: item[0],
            reverse=True,
        ):
            experience = gold_tubes * 5000 + purple_tubes * 2000
            percentage = (count / total_count * 100) if total_count > 0 else 0
            combinations.append(
                DropCombination(
                    claim_count=1,
                    gold_tubes=gold_tubes,
                    purple_tubes=purple_tubes,
                    experience=experience,
                    count=count,
                    percentage=round(percentage, 1),
                )
            )

        level_stats.append(
            SolaLevelStats(
                sola_level=level,
                combinations=combinations,
                total_count=total_count,
                avg_experience=round(avg_exp, 0),
            )
        )

    return DetailedStatsResponse(level_stats=level_stats)


@router.get("/player-ids", response_model=list[str], tags=["tacet"])
def get_player_ids(
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    """获取所有不重复的玩家ID列表"""
    player_ids = db.query(Record.player_id).distinct().all()
    return [pid[0] for pid in player_ids]


@router.delete("/tacet_records/{record_id}", tags=["tacet"])
def delete_record(
    record_id: int,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    """删除指定记录"""
    record = db.query(Record).filter(Record.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="记录不存在")

    db.delete(record)
    db.commit()
    return {"message": "删除成功"}


@router.post("/ascension-records", response_model=list[AscensionRecordResponse], tags=["ascension"])
def create_ascension_records(
    batch: AscensionRecordBatchCreate,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    db_records = []
    for record_data in batch.ascension_records:
        db_record = AscensionRecord(**record_data.model_dump())
        db_records.append(db_record)

    db.add_all(db_records)
    db.commit()

    for record in db_records:
        db.refresh(record)

    return db_records


@router.get("/ascension-records", tags=["ascension"])
def get_ascension_records(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    sola_level: Optional[int] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=1000),
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    query = db.query(AscensionRecord)

    if player_id:
        query = query.filter(AscensionRecord.player_id == player_id)
    if start_date:
        query = query.filter(AscensionRecord.date >= start_date)
    if end_date:
        query = query.filter(AscensionRecord.date <= end_date)
    if sola_level:
        query = query.filter(AscensionRecord.sola_level == sola_level)

    total = query.count()
    records = query.order_by(AscensionRecord.created_at.desc(), AscensionRecord.id.desc()).offset(skip).limit(limit).all()

    return {
        "data": records,
        "total": total,
        "page_size": limit,
        "current_page": skip // limit + 1,
    }


@router.get("/ascension-detailed-stats", response_model=AscensionDetailedStatsResponse, tags=["ascension"])
def get_ascension_detailed_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    query = db.query(AscensionRecord)

    if player_id:
        query = query.filter(AscensionRecord.player_id == player_id)
    if start_date:
        query = query.filter(AscensionRecord.date >= start_date)
    if end_date:
        query = query.filter(AscensionRecord.date <= end_date)

    grouped_stats = query.with_entities(
        AscensionRecord.sola_level,
        AscensionRecord.drop_count,
        func.count().label("count"),
    ).group_by(AscensionRecord.sola_level, AscensionRecord.drop_count).all()

    level_data = {}
    for stat in grouped_stats:
        level = stat.sola_level
        if level not in level_data:
            level_data[level] = []

        level_data[level].append(
            {
                "drop_count": stat.drop_count,
                "count": stat.count,
            }
        )

    level_stats = []
    for level in sorted(level_data.keys(), reverse=True):
        combinations_data = level_data[level]
        total_count = sum(c["count"] for c in combinations_data)
        total_drop_count = sum(c["drop_count"] * c["count"] for c in combinations_data)
        avg_drop_count = total_drop_count / total_count if total_count > 0 else 0

        combinations = []
        for combo in sorted(combinations_data, key=lambda x: x["drop_count"], reverse=True):
            percentage = (combo["count"] / total_count * 100) if total_count > 0 else 0
            combinations.append(
                AscensionDropCombination(
                    drop_count=combo["drop_count"],
                    count=combo["count"],
                    percentage=round(percentage, 1),
                )
            )

        level_stats.append(
            AscensionSolaLevelStats(
                sola_level=level,
                combinations=combinations,
                total_count=total_count,
                avg_drop_count=round(avg_drop_count, 2),
            )
        )

    return AscensionDetailedStatsResponse(level_stats=level_stats)


@router.get("/ascension-player-ids", response_model=list[str], tags=["ascension"])
def get_ascension_player_ids(
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    player_ids = db.query(AscensionRecord.player_id).distinct().all()
    return [pid[0] for pid in player_ids]


@router.delete("/ascension-records/{record_id}", tags=["ascension"])
def delete_ascension_record(
    record_id: int,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    record = db.query(AscensionRecord).filter(AscensionRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="记录不存在")

    db.delete(record)
    db.commit()
    return {"message": "删除成功"}


@router.post("/resonance-records", response_model=list[ResonanceRecordResponse], tags=["resonance"])
def create_resonance_records(
    batch: ResonanceRecordBatchCreate,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    db_records = []
    for record_data in batch.resonance_records:
        db_record = ResonanceRecord(**record_data.model_dump())
        db_records.append(db_record)

    db.add_all(db_records)
    db.commit()

    for record in db_records:
        db.refresh(record)

    return db_records


@router.get("/resonance-records", tags=["resonance"])
def get_resonance_records(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    sola_level: Optional[int] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=1000),
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    query = db.query(ResonanceRecord)

    if player_id:
        query = query.filter(ResonanceRecord.player_id == player_id)
    if start_date:
        query = query.filter(ResonanceRecord.date >= start_date)
    if end_date:
        query = query.filter(ResonanceRecord.date <= end_date)
    if sola_level:
        query = query.filter(ResonanceRecord.sola_level == sola_level)

    total = query.count()
    records = query.order_by(ResonanceRecord.created_at.desc(), ResonanceRecord.id.desc()).offset(skip).limit(limit).all()

    return {
        "data": records,
        "total": total,
        "page_size": limit,
        "current_page": skip // limit + 1,
    }


@router.get("/resonance-detailed-stats", response_model=ResonanceDetailedStatsResponse, tags=["resonance"])
def get_resonance_detailed_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    query = db.query(ResonanceRecord)

    if player_id:
        query = query.filter(ResonanceRecord.player_id == player_id)
    if start_date:
        query = query.filter(ResonanceRecord.date >= start_date)
    if end_date:
        query = query.filter(ResonanceRecord.date <= end_date)

    grouped_stats = query.with_entities(
        ResonanceRecord.sola_level,
        ResonanceRecord.gold,
        ResonanceRecord.purple,
        ResonanceRecord.blue,
        ResonanceRecord.green,
        func.count().label("count"),
    ).group_by(
        ResonanceRecord.sola_level,
        ResonanceRecord.gold,
        ResonanceRecord.purple,
        ResonanceRecord.blue,
        ResonanceRecord.green,
    ).all()

    level_data = {}
    for stat in grouped_stats:
        level = stat.sola_level
        if level not in level_data:
            level_data[level] = []

        level_data[level].append(
            {
                "gold": stat.gold,
                "purple": stat.purple,
                "blue": stat.blue,
                "green": stat.green,
                "count": stat.count,
            }
        )

    level_stats = []
    for level in sorted(level_data.keys(), reverse=True):
        combinations_data = level_data[level]
        total_count = sum(c["count"] for c in combinations_data)

        total_gold = sum(c["gold"] * c["count"] for c in combinations_data)
        total_purple = sum(c["purple"] * c["count"] for c in combinations_data)
        total_blue = sum(c["blue"] * c["count"] for c in combinations_data)
        total_green = sum(c["green"] * c["count"] for c in combinations_data)

        avg_gold = total_gold / total_count if total_count > 0 else 0
        avg_purple = total_purple / total_count if total_count > 0 else 0
        avg_blue = total_blue / total_count if total_count > 0 else 0
        avg_green = total_green / total_count if total_count > 0 else 0

        combinations = []
        for combo in sorted(
            combinations_data,
            key=lambda x: (x["gold"], x["purple"], x["blue"], x["green"]),
            reverse=True,
        ):
            percentage = (combo["count"] / total_count * 100) if total_count > 0 else 0
            combinations.append(
                ResonanceDropCombination(
                    gold=combo["gold"],
                    purple=combo["purple"],
                    blue=combo["blue"],
                    green=combo["green"],
                    count=combo["count"],
                    percentage=round(percentage, 1),
                )
            )

        level_stats.append(
            ResonanceSolaLevelStats(
                sola_level=level,
                combinations=combinations,
                total_count=total_count,
                avg_gold=round(avg_gold, 2),
                avg_purple=round(avg_purple, 2),
                avg_blue=round(avg_blue, 2),
                avg_green=round(avg_green, 2),
            )
        )

    return ResonanceDetailedStatsResponse(level_stats=level_stats)


@router.get("/resonance-player-ids", response_model=list[str], tags=["resonance"])
def get_resonance_player_ids(
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_view_permission),
):
    player_ids = db.query(ResonanceRecord.player_id).distinct().all()
    return [pid[0] for pid in player_ids]


@router.delete("/resonance-records/{record_id}", tags=["resonance"])
def delete_resonance_record(
    record_id: int,
    db: Session = Depends(get_db),
    _: list[str] = Depends(require_edit_permission),
):
    record = db.query(ResonanceRecord).filter(ResonanceRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="记录不存在")

    db.delete(record)
    db.commit()
    return {"message": "删除成功"}


@router.get("/auth/me", tags=["auth"])
async def get_auth_me(permissions: list[str] = Depends(require_view_permission)):
    return {"permissions": permissions}
