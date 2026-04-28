from dataclasses import dataclass
from datetime import datetime


@dataclass
class Article:
    url: str
    title: str
    body: str
    published_at: datetime | None
    source: str


@dataclass
class ScenarioSignal:
    scenario_id: int       # 1-12 matching enum
    scenario_name: str
    confidence: float
    source_url: str
    reasoning: str
    timestamp: int         # unix epoch
