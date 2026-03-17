from pydantic import BaseModel, Field
from datetime import date, datetime
from typing import Optional


class RecordBase(BaseModel):
    date: date
    player_id: str
    gold_tubes: int = Field(ge=0, description="金色密音筒数量")
    purple_tubes: int = Field(ge=0, description="紫色密音筒数量")
    sola_level: int = Field(default=8, ge=1, description="索拉等级")


class RecordCreate(RecordBase):
    pass


class RecordBatchCreate(BaseModel):
    records: list[RecordCreate]


class RecordResponse(RecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class StatsResponse(BaseModel):
    total_records: int
    total_gold_tubes: int
    total_purple_tubes: int
    avg_gold_tubes: float
    avg_purple_tubes: float
    player_count: int


class DropCombination(BaseModel):
    """密音筒产出组合"""
    gold_tubes: int
    purple_tubes: int
    experience: int
    count: int
    percentage: float


class SolaLevelStats(BaseModel):
    """索拉等级统计"""
    sola_level: int
    combinations: list[DropCombination]
    total_count: int
    avg_experience: float


class DetailedStatsResponse(BaseModel):
    """详细统计响应"""
    level_stats: list[SolaLevelStats]
