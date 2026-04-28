import feedparser
import httpx
from datetime import datetime
from email.utils import parsedate_to_datetime
from app.models import Article
from app.scraper.base import BaseScraper

FEED_URL = "https://feeds.reuters.com/reuters/technologyNews"


class ReutersScraper(BaseScraper):
    async def fetch(self) -> list[Article]:
        try:
            async with httpx.AsyncClient(timeout=15.0) as client:
                resp = await client.get(FEED_URL)
                resp.raise_for_status()
                raw_xml = resp.text
        except Exception as e:
            print(f"ReutersScraper fetch error: {e}")
            return []

        feed = feedparser.parse(raw_xml)
        articles: list[Article] = []
        for entry in feed.entries[:20]:
            url = entry.get("link", "")
            title = entry.get("title", "")
            body = entry.get("summary", entry.get("description", ""))
            published_at: datetime | None = None
            raw_date = entry.get("published") or entry.get("updated")
            if raw_date:
                try:
                    published_at = parsedate_to_datetime(raw_date)
                except Exception:
                    pass
            if url:
                articles.append(
                    Article(
                        url=url,
                        title=title,
                        body=body,
                        published_at=published_at,
                        source="reuters",
                    )
                )
        return articles
