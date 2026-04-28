from datetime import datetime, timezone

import pytest

from app.models import Article
from app.scraper import runner


class FakeScraper:
    def __init__(self, articles=None, error=None):
        self.articles = articles or []
        self.error = error

    async def fetch(self):
        if self.error:
            raise self.error
        return self.articles


class FakeAcquire:
    def __init__(self, conn):
        self.conn = conn

    async def __aenter__(self):
        return self.conn

    async def __aexit__(self, exc_type, exc, tb):
        return False


class FakeConn:
    def __init__(self, rows):
        self.rows = list(rows)
        self.calls = []

    async def fetchrow(self, query, *args):
        self.calls.append((query, args))
        return self.rows.pop(0) if self.rows else None


class FakePool:
    def __init__(self, conn):
        self.conn = conn

    def acquire(self):
        return FakeAcquire(self.conn)


@pytest.mark.asyncio
async def test_run_all_scrapers_continues_after_scraper_failure(monkeypatch):
    article = Article(
        url="https://example.com/ai",
        title="AI safety update",
        body="body",
        published_at=datetime(2026, 4, 28, tzinfo=timezone.utc),
        source="test",
    )
    monkeypatch.setattr(
        runner,
        "SCRAPERS",
        [FakeScraper(error=RuntimeError("boom")), FakeScraper(articles=[article])],
    )

    articles = await runner.run_all_scrapers()

    assert articles == [article]


@pytest.mark.asyncio
async def test_store_articles_returns_inserted_ids(monkeypatch):
    article = Article(
        url="https://example.com/ai",
        title="AI safety update",
        body="body",
        published_at=datetime(2026, 4, 28, tzinfo=timezone.utc),
        source="test",
    )
    conn = FakeConn([{"id": "article-id"}, None])

    async def fake_get_pool():
        return FakePool(conn)

    monkeypatch.setattr(runner, "get_pool", fake_get_pool)

    stored_ids = await runner.store_articles([article, article])

    assert stored_ids == ["article-id"]
    assert len(conn.calls) == 2
