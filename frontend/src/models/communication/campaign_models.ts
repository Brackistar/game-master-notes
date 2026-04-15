export type CampaignViewModel = {
  id: string;
  planeId: string;
  name: string;
  version: number;
};

export type WorldStatus = "draft" | "active" | "archived";

export type PlaneViewModel = {
  id: string;
  name: string;
  description: string;
  version: number;
};

export type CreateCampaignInput = {
  name: string;
  planeId?: string;
};

export type CreatePlaneInput = {
  name: string;
  description: string;
};

export type UpdatePlaneInput = {
  id: string;
  name: string;
  description: string;
};

export type WorldViewModel = {
  id: string;
  planeId: string;
  name: string;
  description: string;
  status: WorldStatus;
  version: number;
};

export type NoteOwnerType = "campaign" | "player" | "plane";

export type CampaignNoteSummary = {
  id: string;
  title: string;
  noteType: string;
  ownerType: NoteOwnerType;
  ownerId: string;
  ownerName: string;
};

export type CampaignNoteDetail = {
  id: string;
  title: string;
  contentMd: string;
  noteType: string;
  ownerType: NoteOwnerType;
  ownerId: string;
  ownerName: string;
};

export type UpdateCampaignNoteInput = {
  title: string;
  contentMd: string;
  noteType: string;
};

export type CreateNoteInput = {
  title: string;
  contentMd: string;
  noteType: string;
  ownerType: NoteOwnerType;
  ownerId: string;
  ownerName: string;
};

export type CampaignPlaneNote = {
  id: string;
  planeId: string;
  title: string;
  contentMd: string;
};

export type CampaignPlayerSummary = {
  id: string;
  name: string;
};

export type CampaignPlayerDetail = {
  id: string;
  name: string;
};

export type CampaignPlaneSummary = {
  id: string;
  name: string;
  description?: string;
};

export type PlaneWorldSummary = {
  id: string;
  name: string;
  description?: string;
};

export type PlayerNoteCard = {
  id: string;
  title: string;
  noteType: string;
  contentMd: string;
};
