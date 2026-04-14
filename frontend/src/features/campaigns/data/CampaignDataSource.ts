import type {
  CampaignViewModel,
  CreateCampaignInput
} from "../model";

export interface CampaignDataSource {
  listCampaigns(): Promise<CampaignViewModel[]>;
  createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel>;
  deleteCampaign(id: string): Promise<void>;
  reorderCampaigns(idsInOrder: string[]): Promise<void>;
}
