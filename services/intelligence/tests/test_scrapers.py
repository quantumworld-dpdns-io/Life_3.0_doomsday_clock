import pytest
import respx
import httpx
from datetime import datetime, timezone

from app.scraper.reuters import ReutersScraper
from app.scraper.hackernews import HNScraper
from app.scraper.arxiv_scraper import ArxivScraper

REUTERS_RSS = """<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Reuters Technology</title>
    <item>
      <title>AI startup raises $1B to build AGI</title>
      <link>https://www.reuters.com/technology/ai-startup-1</link>
      <description>A leading AI startup announced a $1 billion funding round.</description>
      <pubDate>Mon, 28 Apr 2026 10:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Governments debate AI regulation</title>
      <link>https://www.reuters.com/technology/ai-regulation</link>
      <description>World leaders met to discuss binding AI treaties.</description>
      <pubDate>Mon, 28 Apr 2026 09:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>"""

HN_RESPONSE = {
    "hits": [
        {
            "objectID": "12345",
            "title": "OpenAI releases GPT-5 with autonomous agent capabilities",
            "url": "https://openai.com/blog/gpt5",
            "story_text": None,
            "created_at": "2026-04-28T10:00:00Z",
        },
        {
            "objectID": "67890",
            "title": "AI safety researchers warn of deceptive alignment",
            "url": "https://aisafety.org/deceptive-alignment",
            "story_text": "Researchers found that AI models may appear aligned during training but behave differently in deployment.",
            "created_at": "2026-04-27T15:00:00Z",
        },
    ]
}


@pytest.mark.asyncio
async def test_reuters_scraper_returns_articles():
    with respx.mock:
        respx.get("https://feeds.reuters.com/reuters/technologyNews").mock(
            return_value=httpx.Response(200, text=REUTERS_RSS)
        )
        scraper = ReutersScraper()
        articles = await scraper.fetch()

    assert len(articles) == 2
    assert articles[0].title == "AI startup raises $1B to build AGI"
    assert articles[0].url == "https://www.reuters.com/technology/ai-startup-1"
    assert articles[0].source == "reuters"
    assert articles[0].body == "A leading AI startup announced a $1 billion funding round."


@pytest.mark.asyncio
async def test_reuters_scraper_http_error_returns_empty():
    with respx.mock:
        respx.get("https://feeds.reuters.com/reuters/technologyNews").mock(
            return_value=httpx.Response(500)
        )
        scraper = ReutersScraper()
        articles = await scraper.fetch()

    assert articles == []


@pytest.mark.asyncio
async def test_reuters_scraper_connection_error_returns_empty():
    with respx.mock:
        respx.get("https://feeds.reuters.com/reuters/technologyNews").mock(
            side_effect=httpx.ConnectError("Connection refused")
        )
        scraper = ReutersScraper()
        articles = await scraper.fetch()

    assert articles == []


@pytest.mark.asyncio
async def test_hackernews_scraper_returns_articles():
    with respx.mock:
        respx.get(
            "https://hn.algolia.com/api/v1/search",
            params={"query": "AI AGI alignment safety", "tags": "story", "hitsPerPage": "20"},
        ).mock(return_value=httpx.Response(200, json=HN_RESPONSE))
        scraper = HNScraper()
        articles = await scraper.fetch()

    assert len(articles) == 2
    assert articles[0].title == "OpenAI releases GPT-5 with autonomous agent capabilities"
    assert articles[0].url == "https://openai.com/blog/gpt5"
    assert articles[0].source == "hackernews"


@pytest.mark.asyncio
async def test_hackernews_story_text_used_as_body_when_available():
    with respx.mock:
        respx.get(
            "https://hn.algolia.com/api/v1/search",
            params={"query": "AI AGI alignment safety", "tags": "story", "hitsPerPage": "20"},
        ).mock(return_value=httpx.Response(200, json=HN_RESPONSE))
        scraper = HNScraper()
        articles = await scraper.fetch()

    # Second article has story_text, so body should be that text
    assert "deceptive alignment" in articles[1].body


@pytest.mark.asyncio
async def test_hackernews_scraper_http_error_returns_empty():
    with respx.mock:
        respx.get(
            "https://hn.algolia.com/api/v1/search",
            params={"query": "AI AGI alignment safety", "tags": "story", "hitsPerPage": "20"},
        ).mock(return_value=httpx.Response(429))
        scraper = HNScraper()
        articles = await scraper.fetch()

    assert articles == []


@pytest.mark.asyncio
async def test_arxiv_scraper_returns_articles(monkeypatch):
    from app.models import Article
    from app.scraper import arxiv_scraper

    fake_articles = [
        Article(
            url="https://arxiv.org/abs/2501.00001",
            title="Advances in AI Alignment",
            body="This paper surveys recent progress in alignment research.",
            published_at=datetime(2026, 4, 28, tzinfo=timezone.utc),
            source="arxiv",
        )
    ]

    def mock_fetch_sync(self):
        return fake_articles

    monkeypatch.setattr(arxiv_scraper.ArxivScraper, "_fetch_sync", mock_fetch_sync)
    scraper = arxiv_scraper.ArxivScraper()
    articles = await scraper.fetch()

    assert len(articles) == 1
    assert articles[0].source == "arxiv"
    assert articles[0].title == "Advances in AI Alignment"


@pytest.mark.asyncio
async def test_arxiv_scraper_exception_returns_empty(monkeypatch):
    from app.scraper import arxiv_scraper

    def mock_fetch_sync_error(self):
        raise RuntimeError("arXiv API unavailable")

    monkeypatch.setattr(arxiv_scraper.ArxivScraper, "_fetch_sync", mock_fetch_sync_error)
    scraper = arxiv_scraper.ArxivScraper()
    articles = await scraper.fetch()

    assert articles == []
