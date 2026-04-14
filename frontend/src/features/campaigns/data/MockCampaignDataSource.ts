import type {
  CampaignViewModel,
  CreateCampaignInput
} from "../model";
import type { CampaignDataSource } from "./CampaignDataSource";

const defaultWorldId = "01HQ8J5QK8W8S6W2A3E0P8W001";

const seedCampaigns: CampaignViewModel[] = [
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A001",
    worldId: defaultWorldId,
    name: "Ashes of the Ivory Coast",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A002",
    worldId: defaultWorldId,
    name: "The Glass Crown Conspiracy",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A003",
    worldId: defaultWorldId,
    name: "Shadows Beneath Iron Harbor",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A004",
    worldId: defaultWorldId,
    name: "Stars Over Thornwatch Bastion",
    version: 1
  }
];

export class MockCampaignDataSource implements CampaignDataSource {
  private campaigns: CampaignViewModel[];

  constructor(initial: CampaignViewModel[] = seedCampaigns) {
    this.campaigns = initial.map((item) => ({ ...item }));
  }

  async listCampaigns(): Promise<CampaignViewModel[]> {
    return this.campaigns.map((item) => ({ ...item }));
  }

  async createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel> {
    const created: CampaignViewModel = {
      id: this.generateId(),
      worldId: input.worldId ?? defaultWorldId,
      name: input.name,
      version: 1
    };
    this.campaigns = [created, ...this.campaigns];
    return { ...created };
  }

  async deleteCampaign(id: string): Promise<void> {
    this.campaigns = this.campaigns.filter((item) => item.id !== id);
  }

  async reorderCampaigns(idsInOrder: string[]): Promise<void> {
    const indexById = new Map(idsInOrder.map((id, index) => [id, index]));
    this.campaigns = [...this.campaigns].sort((a, b) => {
      const aIdx = indexById.get(a.id) ?? Number.MAX_SAFE_INTEGER;
      const bIdx = indexById.get(b.id) ?? Number.MAX_SAFE_INTEGER;
      return aIdx - bIdx;
    });
  }

  private generateId(): string {
    const rand = Math.random().toString(36).slice(2, 12).toUpperCase();
    const ts = Date.now().toString(36).toUpperCase();
    return `MOCK${ts}${rand}`.slice(0, 26);
  }
}
