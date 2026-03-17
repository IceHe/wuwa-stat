from sqlalchemy import Column, Integer, String, Date, DateTime
from sqlalchemy.sql import func
from app.database import Base


class Record(Base):
    __tablename__ = "records"

    id = Column(Integer, primary_key=True, index=True)
    date = Column(Date, nullable=False, index=True)
    player_id = Column(String, nullable=False, index=True)
    gold_tubes = Column(Integer, nullable=False, default=0)
    purple_tubes = Column(Integer, nullable=False, default=0)
    sola_level = Column(Integer, nullable=False, default=8)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    def __repr__(self):
        return f"<Record(player_id={self.player_id}, date={self.date})>"
