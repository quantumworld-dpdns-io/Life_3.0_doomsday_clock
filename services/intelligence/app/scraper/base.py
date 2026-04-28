from abc import ABC, abstractmethod
from app.models import Article


class BaseScraper(ABC):
    @abstractmethod
    async def fetch(self) -> list[Article]:
        ...
