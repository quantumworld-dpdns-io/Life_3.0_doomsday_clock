import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from app.models import ScenarioSignal


@pytest.mark.asyncio
async def test_classifier_returns_signal_on_valid_json():
    valid_result = {
        "scenario_id": 12,
        "scenario_name": "Self-Destruction",
        "confidence": 0.92,
        "reasoning": "Autonomous AI disabling infrastructure represents existential threat.",
    }

    with patch("app.classifier.llm.LLMClassifier.__init__", return_value=None):
        from app.classifier.llm import LLMClassifier

        classifier = LLMClassifier.__new__(LLMClassifier)
        classifier._chain = MagicMock()
        classifier._chain.ainvoke = AsyncMock(return_value=valid_result)

        signal = await classifier.classify(
            title="Rogue AI disables power grids across Europe",
            body="An autonomous AI agent has disabled critical power infrastructure.",
            source_url="https://example.com/article/1",
        )

    assert signal is not None
    assert isinstance(signal, ScenarioSignal)
    assert signal.scenario_id == 12
    assert signal.scenario_name == "Self-Destruction"
    assert signal.confidence == pytest.approx(0.92)
    assert signal.source_url == "https://example.com/article/1"
    assert "existential" in signal.reasoning


@pytest.mark.asyncio
async def test_classifier_returns_none_on_invalid_json():
    with patch("app.classifier.llm.LLMClassifier.__init__", return_value=None):
        from app.classifier.llm import LLMClassifier

        classifier = LLMClassifier.__new__(LLMClassifier)
        classifier._chain = MagicMock()
        # Simulate LLM returning a non-dict (parse error)
        classifier._chain.ainvoke = AsyncMock(side_effect=Exception("JSON parse error"))

        signal = await classifier.classify(
            title="Some article",
            body="Some body text",
            source_url="https://example.com/article/2",
        )

    assert signal is None


@pytest.mark.asyncio
async def test_classifier_clamps_confidence_within_bounds():
    valid_result = {
        "scenario_id": 4,
        "scenario_name": "Gatekeeper",
        "confidence": 1.5,  # out of bounds — should be clamped to 1.0
        "reasoning": "Heavy regulatory framework imposed on AI labs.",
    }

    with patch("app.classifier.llm.LLMClassifier.__init__", return_value=None):
        from app.classifier.llm import LLMClassifier

        classifier = LLMClassifier.__new__(LLMClassifier)
        classifier._chain = MagicMock()
        classifier._chain.ainvoke = AsyncMock(return_value=valid_result)

        signal = await classifier.classify(
            title="G7 passes binding AI treaty",
            body="All frontier AI models must pass international safety review.",
            source_url="https://example.com/article/3",
        )

    assert signal is not None
    assert signal.confidence <= 1.0


@pytest.mark.asyncio
async def test_classifier_returns_none_on_invalid_scenario_id():
    bad_result = {
        "scenario_id": 99,  # invalid
        "scenario_name": "Unknown",
        "confidence": 0.5,
        "reasoning": "Some reasoning.",
    }

    with patch("app.classifier.llm.LLMClassifier.__init__", return_value=None):
        from app.classifier.llm import LLMClassifier

        classifier = LLMClassifier.__new__(LLMClassifier)
        classifier._chain = MagicMock()
        classifier._chain.ainvoke = AsyncMock(return_value=bad_result)

        signal = await classifier.classify(
            title="Random article",
            body="Random body",
            source_url="https://example.com/article/4",
        )

    assert signal is None
