import type {
  CampaignNoteDetail,
  CampaignNoteSummary,
  CampaignPlayerDetail,
  CampaignPlayerSummary,
  CreateNoteInput,
  CreatePlaneInput,
  PlayerNoteCard,
  CampaignPlaneSummary,
  CampaignPlaneNote,
  PlaneWorldSummary,
  CampaignViewModel,
  CreateCampaignInput,
  UpdatePlaneInput,
  PlaneViewModel,
  UpdateCampaignNoteInput
} from "../../../models/communication/campaign_models";

export interface CampaignDataSource {
  listPlanes(): Promise<PlaneViewModel[]>;
  createPlane(input: CreatePlaneInput): Promise<PlaneViewModel>;
  updatePlane(input: UpdatePlaneInput): Promise<PlaneViewModel>;
  deletePlane(id: string): Promise<void>;
  listCampaigns(): Promise<CampaignViewModel[]>;
  createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel>;
  deleteCampaign(id: string): Promise<void>;
  reorderCampaigns(idsInOrder: string[]): Promise<void>;
  listPlaneNotesByCampaign(campaignId: string): Promise<CampaignPlaneNote[]>;
  listNotesByCampaign(campaignId: string): Promise<CampaignNoteSummary[]>;
  getNoteById(noteId: string): Promise<CampaignNoteDetail>;
  listPlayersByCampaign(campaignId: string): Promise<CampaignPlayerSummary[]>;
  getPlayerById(playerId: string): Promise<CampaignPlayerDetail>;
  listNotesByPlayer(playerId: string): Promise<PlayerNoteCard[]>;
  listNotesByPlane(planeId: string): Promise<CampaignNoteSummary[]>;
  listPlanesByCampaign(campaignId: string): Promise<CampaignPlaneSummary[]>;
  listWorldsByPlane(planeId: string): Promise<PlaneWorldSummary[]>;
  createNote(input: CreateNoteInput): Promise<CampaignNoteDetail>;
  updateCampaignNote(
    noteId: string,
    input: UpdateCampaignNoteInput
  ): Promise<CampaignNoteDetail>;
}
