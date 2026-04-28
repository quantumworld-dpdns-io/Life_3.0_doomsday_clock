import asyncio
import arxiv
from datetime import timezone
from app.models import Article
from app.scraper.base import BaseScraper

QUERY = 'cat:cs.AI AND (AGI OR "artificial general intelligence" OR "AI safety" OR "alignment")'


class ArxivScraper(BaseScraper):
    async def fetch(self) -> list[Article]:
        try:
            # arxiv library is synchronous; run in executor to avoid blocking event loop
            loop = asyncio.get_event_loop()
            results = await loop.run_in_executor(None, self._fetch_sync)
            return results
        except Exception as e:
            print(f"ArxivScraper fetch error: {e}")
            return []

    def _fetch_sync(self) -> list[Article]:
        client = arxiv.Client(page_size=10, delay_seconds=1)
        search = arxiv.Search(query=QUERY, max_results=10, sort_by=arxiv.SortCriterion.SubmittedDate)
        articles = []
        for result in client.results(search):
            published_at = result.published
            if published_at and published_at.tzinfo is None:
                published_at = published_at.replace(tzinfo=timezone.utc)
            articles.append(Article(
                url=result.entry_id,
                title=result.title,
                body=result.summary,
                published_at=published_at,
                source="arxiv",
            ))
        return articles
