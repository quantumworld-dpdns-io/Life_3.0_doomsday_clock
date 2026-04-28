import time
from pathlib import Path

import yaml
from langchain_core.output_parsers import JsonOutputParser
from langchain_core.prompts import ChatPromptTemplate

from app.config import settings
from app.models import ScenarioSignal

SCENARIOS = {
    1: "Libertarian Utopia",
    2: "Benevolent Dictator",
    3: "Egalitarian Utopia",
    4: "Gatekeeper",
    5: "Protector God",
    6: "Enslaved God",
    7: "Conquerors",
    8: "Descendants",
    9: "Zookeeper",
    10: "1984",
    11: "Revert",
    12: "Self-Destruction",
}
SCENARIO_NAME_TO_ID = {v: k for k, v in SCENARIOS.items()}


def _load_few_shots() -> str:
    path = Path(__file__).parent / "prompts" / "few_shot.yaml"
    data = yaml.safe_load(path.read_text())
    lines = []
    for s in data["scenarios"]:
        lines.append(f"Scenario {s['id']}. {s['scenario']}: {s['description']}")
        lines.append(f"  Example headline: \"{s['example_headline']}\"")
        lines.append(f"  Why: {s['example_reasoning']}")
    return "\n".join(lines)


_FEW_SHOTS = _load_few_shots()

SYSTEM_PROMPT = (
    "You are an AI risk analyst specialising in Max Tegmark's Life 3.0 framework.\n"
    "Given a news article, classify it into exactly one of the 12 Life 3.0 AI evolution scenarios.\n"
    'Return ONLY valid JSON: {{"scenario_id": <1-12>, "scenario_name": "<name>", '
    '"confidence": <0.0-1.0>, "reasoning": "<one sentence>"}}\n\n'
    "Scenarios and examples:\n{few_shots}"
)


class LLMClassifier:
    def __init__(self):
        if settings.use_ollama:
            from langchain_community.llms import Ollama
            llm = Ollama(base_url=settings.ollama_base_url, model="llama3")
        else:
            from langchain_openai import ChatOpenAI
            llm = ChatOpenAI(
                model="gpt-4o-mini",
                api_key=settings.openai_api_key,
                temperature=0.1,
            )

        prompt = ChatPromptTemplate.from_messages([
            ("system", SYSTEM_PROMPT.format(few_shots=_FEW_SHOTS)),
            ("human", "Article title: {title}\n\nArticle text: {body}"),
        ])
        self._chain = prompt | llm | JsonOutputParser()

    async def classify(self, title: str, body: str, source_url: str) -> ScenarioSignal | None:
        try:
            result = await self._chain.ainvoke({"title": title, "body": body[:3000]})
            scenario_id = int(result["scenario_id"])
            if scenario_id not in SCENARIOS:
                raise ValueError(f"Invalid scenario_id: {scenario_id}")
            return ScenarioSignal(
                scenario_id=scenario_id,
                scenario_name=result.get("scenario_name", SCENARIOS[scenario_id]),
                confidence=max(0.0, min(1.0, float(result["confidence"]))),
                source_url=source_url,
                reasoning=result.get("reasoning", ""),
                timestamp=int(time.time()),
            )
        except Exception as e:
            print(f"Classification failed for {source_url}: {e}")
            return None
