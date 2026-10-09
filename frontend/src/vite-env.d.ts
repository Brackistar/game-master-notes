/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_CAMPAIGN_DATA_SOURCE?: "mock" | "api";
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_DEFAULT_WORLD_ID?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
