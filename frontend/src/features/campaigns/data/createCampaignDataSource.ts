import { ApiCampaignDataSource } from "./ApiCampaignDataSource";
import type { CampaignDataSource } from "./CampaignDataSource";
import { MockCampaignDataSource } from "./MockCampaignDataSource";

export function createCampaignDataSource(): CampaignDataSource {
  const mode = import.meta.env.VITE_CAMPAIGN_DATA_SOURCE;
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";
  const defaultWorldId =
    import.meta.env.VITE_DEFAULT_WORLD_ID ?? "01HQ8J5QK8W8S6W2A3E0P8W001";

  if (mode === "api" || (!import.meta.env.DEV && mode !== "mock")) {
    return new ApiCampaignDataSource(apiBaseUrl, defaultWorldId);
  }

  return new MockCampaignDataSource();
}
