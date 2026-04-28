import feedparser
import httpx
from datetime import datetime
from email.utils import parsedate_to_datetime
from app.models import Article
from app.scraper.base import BaseScraper

FEED_URL = "https://futureoflife.org/feed/"


class FLIScraper(BaseScraper):
    async def fetch(self) -> list[Article]:
        try:
            async with httpx.AsyncClient(timeout=15.0, follow_redirects=True) as client:
                resp = await client.get(FEED_URL)
                resp.raise_for_status()
                raw_xml = resp.text
        except Exception as e:
            print(f"FLIScraper fetch error: {e}")
            return []

        feed = feedparser.parse(raw_xml)
        articles: list[Article] = []
        for entry in feed.entries[:20]:
            url = entry.get("link", "")
            title = entry.get("title", "")
            # Prefer full content over summary
            content_list = entry.get("content", [])
            if content_list:
                body = content_list[0].get("value", "")
            else:
                body = entry.get("summary", "")
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
                        source="fli",
                    )
                )
        return articles
