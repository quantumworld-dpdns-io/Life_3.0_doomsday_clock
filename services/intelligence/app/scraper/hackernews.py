import httpx
from datetime import datetime

from app.models import Article
from app.scraper.base import BaseScraper

ALGOLIA_URL = (
    "https://hn.algolia.com/api/v1/search"
    "?query=AI+AGI+alignment+safety&tags=story&hitsPerPage=20"
)


class HNScraper(BaseScraper):
    async def fetch(self) -> list[Article]:
        try:
            async with httpx.AsyncClient(timeout=15.0) as client:
                resp = await client.get(ALGOLIA_URL)
                resp.raise_for_status()
                data = resp.json()
        except Exception as e:
            print(f"HNScraper fetch error: {e}")
            return []

        articles: list[Article] = []
        for hit in data.get("hits", []):
            object_id = hit.get("objectID", "")
            url = hit.get("url") or f"https://news.ycombinator.com/item?id={object_id}"
            title = hit.get("title", "")
            body = hit.get("story_text") or title
            created_at_raw = hit.get("created_at")
            published_at: datetime | None = None
            if created_at_raw:
                try:
                    published_at = datetime.fromisoformat(created_at_raw.replace("Z", "+00:00"))
                except Exception:
                    pass
            if url and title:
                articles.append(
                    Article(
                        url=url,
                        title=title,
                        body=body,
                        published_at=published_at,
                        source="hackernews",
                    )
                )
        return articles
