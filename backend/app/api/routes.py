from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import func
from typing import Optional
from datetime import date

from app.database import get_db
from app.models import Record
from app.schemas import (
    RecordCreate,
    RecordBatchCreate,
    RecordResponse,
    StatsResponse,
    DetailedStatsResponse,
    SolaLevelStats,
    DropCombination,
)

router = APIRouter(prefix="/api", tags=["records"])


@router.post("/records", response_model=list[RecordResponse])
def create_records(
    batch: RecordBatchCreate,
    db: Session = Depends(get_db)
):
    """批量创建记录"""
    db_records = []
    for record_data in batch.records:
        db_record = Record(**record_data.model_dump())
        db_records.append(db_record)

    db.add_all(db_records)
    db.commit()

    for record in db_records:
        db.refresh(record)

    return db_records


@router.get("/records", response_model=list[RecordResponse])
def get_records(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    sola_level: Optional[int] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db)
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

    records = query.order_by(Record.date.desc()).offset(skip).limit(limit).all()
    return records


@router.get("/stats", response_model=StatsResponse)
def get_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
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
            total_gold_tubes=0,
            total_purple_tubes=0,
            avg_gold_tubes=0.0,
            avg_purple_tubes=0.0,
            player_count=0
        )

    stats = query.with_entities(
        func.sum(Record.gold_tubes).label("total_gold"),
        func.sum(Record.purple_tubes).label("total_purple"),
        func.avg(Record.gold_tubes).label("avg_gold"),
        func.avg(Record.purple_tubes).label("avg_purple"),
        func.count(func.distinct(Record.player_id)).label("player_count")
    ).first()

    return StatsResponse(
        total_records=total_records,
        total_gold_tubes=stats.total_gold or 0,
        total_purple_tubes=stats.total_purple or 0,
        avg_gold_tubes=float(stats.avg_gold or 0),
        avg_purple_tubes=float(stats.avg_purple or 0),
        player_count=stats.player_count or 0
    )


@router.get("/detailed-stats", response_model=DetailedStatsResponse)
def get_detailed_stats(
    player_id: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    """获取详细统计数据，按索拉等级和密音筒组合分组"""
    query = db.query(Record)

    if player_id:
        query = query.filter(Record.player_id == player_id)
    if start_date:
        query = query.filter(Record.date >= start_date)
    if end_date:
        query = query.filter(Record.date <= end_date)

    # 按索拉等级和密音筒组合分组统计
    grouped_stats = query.with_entities(
        Record.sola_level,
        Record.gold_tubes,
        Record.purple_tubes,
        func.count().label('count')
    ).group_by(
        Record.sola_level,
        Record.gold_tubes,
        Record.purple_tubes
    ).all()

    # 组织数据结构
    level_data = {}
    for stat in grouped_stats:
        level = stat.sola_level
        if level not in level_data:
            level_data[level] = []

        # 计算经验：金色 * 5000 + 紫色 * 2000
        experience = stat.gold_tubes * 5000 + stat.purple_tubes * 2000

        level_data[level].append({
            'gold_tubes': stat.gold_tubes,
            'purple_tubes': stat.purple_tubes,
            'experience': experience,
            'count': stat.count
        })

    # 计算每个等级的统计数据
    level_stats = []
    for level in sorted(level_data.keys(), reverse=True):
        combinations_data = level_data[level]
        total_count = sum(c['count'] for c in combinations_data)

        # 计算平均经验
        total_exp = sum(c['experience'] * c['count'] for c in combinations_data)
        avg_exp = total_exp / total_count if total_count > 0 else 0

        # 计算每种组合的占比
        combinations = []
        for combo in sorted(combinations_data,
                          key=lambda x: (x['gold_tubes'], x['purple_tubes']),
                          reverse=True):
            percentage = (combo['count'] / total_count * 100) if total_count > 0 else 0
            combinations.append(DropCombination(
                gold_tubes=combo['gold_tubes'],
                purple_tubes=combo['purple_tubes'],
                experience=combo['experience'],
                count=combo['count'],
                percentage=round(percentage, 1)
            ))

        level_stats.append(SolaLevelStats(
            sola_level=level,
            combinations=combinations,
            total_count=total_count,
            avg_experience=round(avg_exp, 0)
        ))

    return DetailedStatsResponse(level_stats=level_stats)


@router.get("/player-ids", response_model=list[str])
def get_player_ids(db: Session = Depends(get_db)):
    """获取所有不重复的玩家ID列表"""
    player_ids = db.query(Record.player_id).distinct().all()
    return [pid[0] for pid in player_ids]


@router.delete("/records/{record_id}")
def delete_record(record_id: int, db: Session = Depends(get_db)):
    """删除指定记录"""
    record = db.query(Record).filter(Record.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="记录不存在")

    db.delete(record)
    db.commit()
    return {"message": "删除成功"}
