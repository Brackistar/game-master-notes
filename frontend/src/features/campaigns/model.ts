export type CampaignViewModel = {
  id: string;
  worldId: string;
  name: string;
  version: number;
};

export type CreateCampaignInput = {
  name: string;
  worldId?: string;
};
