from app.scraper.reuters import ReutersScraper
from app.scraper.arxiv_scraper import ArxivScraper
from app.scraper.hackernews import HNScraper
from app.scraper.fli import FLIScraper
from app.db import get_pool
from app.models import Article

SCRAPERS = [ReutersScraper(), ArxivScraper(), HNScraper(), FLIScraper()]


async def run_all_scrapers() -> list[Article]:
    results = []
    for scraper in SCRAPERS:
        try:
            articles = await scraper.fetch()
            results.extend(articles)
        except Exception as e:
            print(f"Scraper {scraper.__class__.__name__} failed: {e}")
    return results


async def store_articles(articles: list[Article]) -> list[str]:
    pool = await get_pool()
    stored_ids = []
    async with pool.acquire() as conn:
        for a in articles:
            try:
                row = await conn.fetchrow(
                    """INSERT INTO raw_articles (url, title, body, published_at, source)
                       VALUES ($1, $2, $3, $4, $5)
                       ON CONFLICT (url) DO NOTHING
                       RETURNING id""",
                    a.url, a.title, a.body, a.published_at, a.source,
                )
                if row:
                    stored_ids.append(str(row["id"]))
            except Exception as e:
                print(f"Failed to store article {a.url}: {e}")
    return stored_ids
