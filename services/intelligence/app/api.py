import asyncio
from contextlib import asynccontextmanager, suppress
from pathlib import Path

from fastapi import FastAPI, Header, HTTPException, Query

from app.config import settings
from app.db import close_pool, get_pool
from app.scheduler import start_scheduler, stop_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    pool = await get_pool()
    migration_sql = (Path(__file__).parent / "migrations" / "001_init.sql").read_text()
    async with pool.acquire() as conn:
        await conn.execute(migration_sql)
    start_scheduler()
    from app.grpc_server import serve as serve_grpc

    grpc_task = asyncio.create_task(serve_grpc(), name="intelligence-grpc")
    try:
        yield
    finally:
        stop_scheduler()
        grpc_task.cancel()
        with suppress(asyncio.CancelledError):
            await grpc_task
        await close_pool()


app = FastAPI(title="Life 3.0 Intelligence Service", lifespan=lifespan)


@app.get("/health")
async def health():
    return {"status": "ok"}


@app.get("/signals/latest")
async def latest_signals(limit: int = Query(50, ge=1, le=100)):
    pool = await get_pool()
    rows = await pool.fetch(
        """
        SELECT s.id, s.scenario, s.confidence, s.reasoning, a.url, s.created_at
        FROM scenario_signals s
        JOIN raw_articles a ON a.id = s.article_id
        ORDER BY s.created_at DESC
        LIMIT $1
        """,
        limit,
    )
    return [
        {
            "id": str(r["id"]),
            "scenario": r["scenario"],
            "confidence": r["confidence"],
            "reasoning": r["reasoning"],
            "url": r["url"],
            "created_at": r["created_at"].isoformat() if r["created_at"] else None,
        }
        for r in rows
    ]


@app.get("/signals/aggregate")
async def aggregate_signals(window_days: int = Query(7, ge=1, le=365)):
    pool = await get_pool()
    rows = await pool.fetch(
        """
        SELECT scenario, AVG(confidence) AS avg_confidence, COUNT(*) AS count
        FROM scenario_signals
        WHERE created_at > NOW() - ($1 * INTERVAL '1 day')
        GROUP BY scenario
        ORDER BY scenario
        """,
        window_days,
    )
    weights = [0.0] * 13  # index 0 unused; 1-12 = scenario id
    for r in rows:
        weights[r["scenario"]] = float(r["avg_confidence"])
    return {"weights": weights, "window_days": window_days}


@app.post("/scrape/trigger")
async def trigger_scrape(x_admin_key: str = Header(...)):
    if x_admin_key != settings.admin_api_key:
        raise HTTPException(status_code=403, detail="Invalid admin key")
    asyncio.create_task(_run_scrape())
    return {"status": "triggered"}


async def _run_scrape():
    from app.scraper.runner import run_all_scrapers, store_articles
    from app.classifier.pipeline import classify_unprocessed

    articles = await run_all_scrapers()
    await store_articles(articles)
    await classify_unprocessed()
