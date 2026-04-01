from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy import inspect, text

from app.api.routes import router
from app.database import Base, engine, settings


def ensure_tacet_record_schema() -> None:
    inspector = inspect(engine)
    if "tacet_records" not in inspector.get_table_names():
        return

    columns = {column["name"] for column in inspector.get_columns("tacet_records")}

    with engine.begin() as connection:
        if "claim_count" not in columns:
            connection.execute(
                text(
                    "ALTER TABLE tacet_records "
                    "ADD COLUMN claim_count INTEGER NOT NULL DEFAULT 1"
                )
            )
            if "reward_mode" in columns:
                connection.execute(
                    text(
                        "UPDATE tacet_records "
                        "SET claim_count = CASE WHEN reward_mode = 'double' THEN 2 ELSE 1 END"
                    )
                )


def initialize_database() -> None:
    Base.metadata.create_all(bind=engine)
    ensure_tacet_record_schema()


initialize_database()

app = FastAPI(
    title="鸣潮产出统计",
    description="用于统计鸣潮的无音区、共鸣者突破材料和凝素领域产出数据",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=[settings.frontend_url],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)


@app.get("/")
def root():
    return {"message": "鸣潮产出统计 API"}


@app.get("/health")
def health_check():
    return {"status": "ok"}
