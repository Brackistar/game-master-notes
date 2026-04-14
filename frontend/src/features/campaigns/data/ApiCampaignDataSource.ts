import type { CampaignDataSource } from "./CampaignDataSource";
import type {
  CampaignViewModel,
  CreateCampaignInput
} from "../model";

type CampaignResponse = {
  id: string;
  world_id: string;
  name: string;
  version: number;
};

type ListCampaignsResponse = {
  items: CampaignResponse[];
};

export class ApiCampaignDataSource implements CampaignDataSource {
  constructor(
    private readonly baseUrl: string,
    private readonly defaultWorldId: string
  ) {}

  async listCampaigns(): Promise<CampaignViewModel[]> {
    const response = await fetch(`${this.baseUrl}/campaigns`);
    if (!response.ok) {
      throw new Error("Unable to load campaigns.");
    }

    const payload = (await response.json()) as ListCampaignsResponse;
    return payload.items.map(mapCampaignResponse);
  }

  async createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel> {
    const response = await fetch(`${this.baseUrl}/campaigns`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        world_id: input.worldId ?? this.defaultWorldId,
        name: input.name
      })
    });
    if (!response.ok) {
      throw new Error("Unable to create campaign.");
    }
    const payload = (await response.json()) as CampaignResponse;
    return mapCampaignResponse(payload);
  }

  async deleteCampaign(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/campaigns/${id}`, {
      method: "DELETE"
    });
    if (!response.ok) {
      throw new Error("Unable to delete campaign.");
    }
  }

  async reorderCampaigns(_: string[]): Promise<void> {
    return Promise.resolve();
  }
}

function mapCampaignResponse(item: CampaignResponse): CampaignViewModel {
  return {
    id: item.id,
    worldId: item.world_id,
    name: item.name,
    version: item.version
  };
}
