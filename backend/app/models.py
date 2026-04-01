from sqlalchemy import Column, Date, DateTime, Integer, String
from sqlalchemy.sql import func

from app.database import Base


class Record(Base):
    __tablename__ = "tacet_records"

    id = Column(Integer, primary_key=True, index=True)
    date = Column(Date, nullable=False, index=True)
    player_id = Column(String, nullable=False, index=True)
    gold_tubes = Column(Integer, nullable=False, default=0)
    purple_tubes = Column(Integer, nullable=False, default=0)
    claim_count = Column(Integer, nullable=False, default=1)
    sola_level = Column(Integer, nullable=False, default=8)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    def __repr__(self) -> str:
        return f"<Record(player_id={self.player_id}, date={self.date})>"


class AscensionRecord(Base):
    __tablename__ = "ascension_records"

    id = Column(Integer, primary_key=True, index=True)
    date = Column(Date, nullable=False, index=True)
    player_id = Column(String, nullable=False, index=True)
    sola_level = Column(Integer, nullable=False, default=8)
    drop_count = Column(Integer, nullable=False, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    def __repr__(self) -> str:
        return f"<AscensionRecord(player_id={self.player_id}, date={self.date})>"


class ResonanceRecord(Base):
    __tablename__ = "resonance_records"

    id = Column(Integer, primary_key=True, index=True)
    date = Column(Date, nullable=False, index=True)
    player_id = Column(String, nullable=False, index=True)
    sola_level = Column(Integer, nullable=False, default=8)
    gold = Column(Integer, nullable=False, default=0)
    purple = Column(Integer, nullable=False, default=0)
    blue = Column(Integer, nullable=False, default=0)
    green = Column(Integer, nullable=False, default=0)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    def __repr__(self) -> str:
        return f"<ResonanceRecord(player_id={self.player_id}, date={self.date})>"
