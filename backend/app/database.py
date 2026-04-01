from collections.abc import Generator
from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict
from sqlalchemy import create_engine
from sqlalchemy.orm import Session, declarative_base, sessionmaker


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env")

    database_url: str
    frontend_url: str = "http://localhost:5173"
    auth_service_url: str = "http://127.0.0.1:8080"
    auth_service_timeout_seconds: float = 3.0


@lru_cache()
def get_settings() -> Settings:
    return Settings()


settings = get_settings()

engine = create_engine(settings.database_url)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()


def get_db() -> Generator[Session, None, None]:
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
