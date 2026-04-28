from datetime import datetime, timezone
from unittest.mock import AsyncMock

import pytest
from fastapi.testclient import TestClient

from app import api

client = TestClient(api.app)


class FakePool:
    def __init__(self, rows):
        self.rows = rows
        self.calls = []

    async def fetch(self, query, *args):
        self.calls.append((query, args))
        return self.rows


@pytest.mark.asyncio
async def test_latest_signals_returns_serialized_rows(monkeypatch):
    created_at = datetime(2026, 4, 28, 10, 0, tzinfo=timezone.utc)
    pool = FakePool([
        {
            "id": "e0d90605-0b2b-4aa2-9867-0eb9227868ff",
            "scenario": 12,
            "confidence": 0.91,
            "reasoning": "Autonomous systems caused a cascading failure.",
            "url": "https://example.com/signal",
            "created_at": created_at,
        }
    ])
    monkeypatch.setattr(api, "get_pool", AsyncMock(return_value=pool))

    response = await api.latest_signals(limit=5)

    assert response == [
        {
            "id": "e0d90605-0b2b-4aa2-9867-0eb9227868ff",
            "scenario": 12,
            "confidence": 0.91,
            "reasoning": "Autonomous systems caused a cascading failure.",
            "url": "https://example.com/signal",
            "created_at": "2026-04-28T10:00:00+00:00",
        }
    ]
    assert pool.calls[0][1] == (5,)


@pytest.mark.asyncio
async def test_aggregate_signals_returns_13_slot_weight_vector(monkeypatch):
    pool = FakePool([
        {"scenario": 4, "avg_confidence": 0.25},
        {"scenario": 12, "avg_confidence": 0.75},
    ])
    monkeypatch.setattr(api, "get_pool", AsyncMock(return_value=pool))

    response = await api.aggregate_signals(window_days=14)

    assert response["window_days"] == 14
    assert len(response["weights"]) == 13
    assert response["weights"][0] == 0.0
    assert response["weights"][4] == 0.25
    assert response["weights"][12] == 0.75


def test_latest_signals_rejects_invalid_limit():
    response = client.get("/signals/latest?limit=0")

    assert response.status_code == 422


def test_aggregate_rejects_invalid_window():
    response = client.get("/signals/aggregate?window_days=0")

    assert response.status_code == 422


def test_trigger_scrape_rejects_bad_admin_key():
    response = client.post("/scrape/trigger", headers={"X-Admin-Key": "wrong"})

    assert response.status_code == 403
