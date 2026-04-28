from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env")

    database_url: str = "postgres://life3:life3@localhost:5432/life3"
    admin_api_key: str = "admin"
    openai_api_key: str = ""
    ollama_base_url: str = ""
    use_ollama: bool = False
    grpc_port: int = 50051
    http_port: int = 8001
    scrape_interval_minutes: int = 15


settings = Settings()
