from app.classifier.llm import LLMClassifier
from app.db import get_pool

_classifier = LLMClassifier()


async def classify_unprocessed():
    pool = await get_pool()
    async with pool.acquire() as conn:
        rows = await conn.fetch("""
            SELECT a.id, a.title, a.body, a.url
            FROM raw_articles a
            WHERE NOT EXISTS (
                SELECT 1 FROM scenario_signals s WHERE s.article_id = a.id
            )
            ORDER BY a.ingested_at DESC
            LIMIT 50
        """)

    for row in rows:
        signal = await _classifier.classify(row["title"] or "", row["body"] or "", row["url"])
        if signal:
            async with pool.acquire() as conn:
                await conn.execute(
                    """INSERT INTO scenario_signals (article_id, scenario, confidence, reasoning)
                       VALUES ($1, $2, $3, $4)
                       ON CONFLICT DO NOTHING""",
                    row["id"],
                    signal.scenario_id,
                    signal.confidence,
                    signal.reasoning,
                )
