import { ApiCampaignDataSource } from "./ApiCampaignDataSource";
import type { CampaignDataSource } from "./CampaignDataSource";
import { MockCampaignDataSource } from "./MockCampaignDataSource";

export function createCampaignDataSource(): CampaignDataSource {
  const mode = import.meta.env.VITE_CAMPAIGN_DATA_SOURCE;
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";
  const defaultPlaneId =
    import.meta.env.VITE_DEFAULT_PLANE_ID ?? "01HQ8J5QK8W8S6W2A3E0P8P001";

  if (mode === "mock" || import.meta.env.MODE === "test") {
    return new MockCampaignDataSource();
  }
  return new ApiCampaignDataSource(apiBaseUrl, defaultPlaneId);
}
