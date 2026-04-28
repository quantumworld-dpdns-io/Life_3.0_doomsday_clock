from apscheduler.schedulers.asyncio import AsyncIOScheduler

from app.config import settings

scheduler = AsyncIOScheduler()


async def _scrape_and_classify():
    from app.scraper.runner import run_all_scrapers, store_articles
    from app.classifier.pipeline import classify_unprocessed

    print("Running scrape cycle...")
    articles = await run_all_scrapers()
    await store_articles(articles)
    await classify_unprocessed()
    print(f"Scrape cycle complete. Processed {len(articles)} articles.")


def start_scheduler():
    scheduler.add_job(
        _scrape_and_classify,
        "interval",
        minutes=settings.scrape_interval_minutes,
        id="scrape",
        replace_existing=True,
    )
    scheduler.start()


def stop_scheduler():
    if scheduler.running:
        scheduler.shutdown()
