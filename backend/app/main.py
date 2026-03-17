from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.database import engine, Base, settings
from app.api.routes import router

# 创建数据库表
Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="鸣潮无音区产出统计",
    description="统计鸣潮游戏无音区的金色和紫色密音筒产出",
    version="1.0.0"
)

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=[settings.frontend_url],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册路由
app.include_router(router)


@app.get("/")
def root():
    return {"message": "鸣潮无音区产出统计 API"}


@app.get("/health")
def health_check():
    return {"status": "ok"}
