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
    tacet_records: list[RecordCreate]


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


class AscensionRecordBase(BaseModel):
    date: date
    player_id: str
    sola_level: int = Field(default=8, ge=1, description="索拉等级")
    drop_count: int = Field(ge=0, description="突破材料掉落数量")


class AscensionRecordCreate(AscensionRecordBase):
    pass


class AscensionRecordBatchCreate(BaseModel):
    ascension_records: list[AscensionRecordCreate]


class AscensionRecordResponse(AscensionRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class AscensionDropCombination(BaseModel):
    drop_count: int
    count: int
    percentage: float


class AscensionSolaLevelStats(BaseModel):
    sola_level: int
    combinations: list[AscensionDropCombination]
    total_count: int
    avg_drop_count: float


class AscensionDetailedStatsResponse(BaseModel):
    level_stats: list[AscensionSolaLevelStats]


class ResonanceRecordBase(BaseModel):
    date: date
    player_id: str
    sola_level: int = Field(default=8, ge=1, description="索拉等级")
    gold: int = Field(ge=0, description="金色掉落数量")
    purple: int = Field(ge=0, description="紫色掉落数量")
    blue: int = Field(ge=0, description="蓝色掉落数量")
    green: int = Field(ge=0, description="绿色掉落数量")


class ResonanceRecordCreate(ResonanceRecordBase):
    pass


class ResonanceRecordBatchCreate(BaseModel):
    resonance_records: list[ResonanceRecordCreate]


class ResonanceRecordResponse(ResonanceRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ResonanceDropCombination(BaseModel):
    gold: int
    purple: int
    blue: int
    green: int
    count: int
    percentage: float


class ResonanceSolaLevelStats(BaseModel):
    sola_level: int
    combinations: list[ResonanceDropCombination]
    total_count: int
    avg_gold: float
    avg_purple: float
    avg_blue: float
    avg_green: float


class ResonanceDetailedStatsResponse(BaseModel):
    level_stats: list[ResonanceSolaLevelStats]
