import type { CampaignDataSource } from "./CampaignDataSource";
import type {
  CampaignNoteDetail,
  CampaignNoteSummary,
  CampaignPlayerDetail,
  CampaignPlayerSummary,
  CampaignPlaneNote,
  CampaignPlaneSummary,
  CampaignViewModel,
  CreateCampaignInput,
  CreateNoteInput,
  CreatePlaneInput,
  PlaneViewModel,
  PlaneWorldSummary,
  PlayerNoteCard,
  UpdateCampaignNoteInput,
  UpdatePlaneInput,
  WorldViewModel
} from "../../../models/communication/campaign_models";

type CampaignResponse = {
  id: string;
  world_id: string;
  name: string;
  version: number;
};

type ListCampaignsResponse = {
  items: CampaignResponse[];
};

type PlaneResponse = {
  id: string;
  name: string;
  description: string;
  version: number;
};

type ListPlanesResponse = {
  items: PlaneResponse[];
};

type WorldResponse = {
  id: string;
  plane_id: string;
  name: string;
  description: string;
  status: string;
  version: number;
};

type ListWorldsResponse = {
  items: WorldResponse[];
};

type NoteOwnerResponse = {
  note_id: string;
};

type ListNoteOwnersResponse = {
  items: NoteOwnerResponse[];
};

type NoteResponse = {
  id: string;
  title: string;
  content_md: string;
  note_type: string;
};

type PlayerResponse = {
  id: string;
  name: string;
};

type CampaignPlayerRelationResponse = {
  player_id: string;
};

type ListCampaignPlayersResponse = {
  items: CampaignPlayerRelationResponse[];
};

export class ApiCampaignDataSource implements CampaignDataSource {
  constructor(
    private readonly baseUrl: string,
    private readonly defaultPlaneId: string
  ) {}

  async listPlanes(): Promise<PlaneViewModel[]> {
    const response = await fetch(`${this.baseUrl}/planes`);
    if (!response.ok) {
      throw new Error("Unable to load planes.");
    }
    const payload = (await response.json()) as ListPlanesResponse;
    return payload.items.map((item) => ({
      id: item.id,
      name: item.name,
      description: item.description,
      version: item.version
    }));
  }

  async createPlane(input: CreatePlaneInput): Promise<PlaneViewModel> {
    const response = await fetch(`${this.baseUrl}/planes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: input.name,
        description: input.description
      })
    });
    if (!response.ok) {
      throw new Error("Unable to create plane.");
    }
    const created = (await response.json()) as PlaneResponse;
    return {
      id: created.id,
      name: created.name,
      description: created.description,
      version: created.version
    };
  }

  async updatePlane(input: UpdatePlaneInput): Promise<PlaneViewModel> {
    const response = await fetch(`${this.baseUrl}/planes/${input.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: input.name,
        description: input.description
      })
    });
    if (!response.ok) {
      throw new Error("Unable to update plane.");
    }
    const updated = (await response.json()) as PlaneResponse;
    return {
      id: updated.id,
      name: updated.name,
      description: updated.description,
      version: updated.version
    };
  }

  async deletePlane(id: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/planes/${id}`, {
      method: "DELETE"
    });
    if (!response.ok) {
      throw new Error("Unable to delete plane.");
    }
  }

  async listCampaigns(): Promise<CampaignViewModel[]> {
    const [campaignResponse, worlds] = await Promise.all([
      fetch(`${this.baseUrl}/campaigns`),
      this.listWorlds()
    ]);
    if (!campaignResponse.ok) {
      throw new Error("Unable to load campaigns.");
    }

    const payload = (await campaignResponse.json()) as ListCampaignsResponse;
    const worldToPlane = new Map(worlds.map((world) => [world.id, world.planeId]));

    return payload.items.map((item) => ({
      id: item.id,
      planeId: worldToPlane.get(item.world_id) ?? this.defaultPlaneId,
      name: item.name,
      version: item.version
    }));
  }

  async createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel> {
    const targetPlaneId = input.planeId ?? this.defaultPlaneId;
    const worldId = await this.resolveWorldIDForPlane(targetPlaneId);

    const response = await fetch(`${this.baseUrl}/campaigns`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        world_id: worldId,
        name: input.name
      })
    });
    if (!response.ok) {
      throw new Error("Unable to create campaign.");
    }
    const payload = (await response.json()) as CampaignResponse;
    return {
      id: payload.id,
      planeId: targetPlaneId,
      name: payload.name,
      version: payload.version
    };
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

  async listPlaneNotesByCampaign(campaignId: string): Promise<CampaignPlaneNote[]> {
    const plane = await this.fetchPlaneForCampaign(campaignId);
    return [
      {
        id: `plane-note-${plane.id}`,
        planeId: plane.id,
        title: plane.name,
        contentMd: plane.description || "No plane description yet."
      }
    ];
  }

  async listNotesByCampaign(campaignId: string): Promise<CampaignNoteSummary[]> {
    const relationsResponse = await fetch(
      `${this.baseUrl}/owners/campaign/${campaignId}/notes`
    );
    if (!relationsResponse.ok) {
      throw new Error("Unable to load campaign notes.");
    }
    const relations = (await relationsResponse.json()) as ListNoteOwnersResponse;
    if (!relations.items.length) {
      return [];
    }

    const notePayloads = await Promise.all(
      relations.items.map(async (relation) => {
        const response = await fetch(`${this.baseUrl}/notes/${relation.note_id}`);
        if (!response.ok) {
          throw new Error("Unable to load note.");
        }
        return (await response.json()) as NoteResponse;
      })
    );

    return notePayloads.map((item) => ({
      id: item.id,
      title: item.title,
      noteType: item.note_type,
      ownerType: "campaign",
      ownerId: campaignId,
      ownerName: ""
    }));
  }

  async getNoteById(noteId: string): Promise<CampaignNoteDetail> {
    const response = await fetch(`${this.baseUrl}/notes/${noteId}`);
    if (!response.ok) {
      throw new Error("Unable to load note.");
    }
    const item = (await response.json()) as NoteResponse;
    return {
      id: item.id,
      title: item.title,
      contentMd: item.content_md,
      noteType: item.note_type,
      ownerType: "campaign",
      ownerId: "",
      ownerName: ""
    };
  }

  async listPlayersByCampaign(campaignId: string): Promise<CampaignPlayerSummary[]> {
    const relationResponse = await fetch(`${this.baseUrl}/campaigns/${campaignId}/players`);
    if (!relationResponse.ok) {
      throw new Error("Unable to load campaign players.");
    }
    const relations = (await relationResponse.json()) as ListCampaignPlayersResponse;
    if (!relations.items.length) {
      return [];
    }

    const players = await Promise.all(
      relations.items.map(async (relation) => {
        const response = await fetch(`${this.baseUrl}/players/${relation.player_id}`);
        if (!response.ok) {
          throw new Error("Unable to load player.");
        }
        return (await response.json()) as PlayerResponse;
      })
    );

    return players.map((player) => ({
      id: player.id,
      name: player.name
    }));
  }

  async getPlayerById(playerId: string): Promise<CampaignPlayerDetail> {
    const response = await fetch(`${this.baseUrl}/players/${playerId}`);
    if (!response.ok) {
      throw new Error("Unable to load player.");
    }
    const player = (await response.json()) as PlayerResponse;
    return {
      id: player.id,
      name: player.name
    };
  }

  async listNotesByPlayer(playerId: string): Promise<PlayerNoteCard[]> {
    const relationsResponse = await fetch(`${this.baseUrl}/owners/player/${playerId}/notes`);
    if (!relationsResponse.ok) {
      throw new Error("Unable to load player notes.");
    }
    const relations = (await relationsResponse.json()) as ListNoteOwnersResponse;
    if (!relations.items.length) {
      return [];
    }

    const notes = await Promise.all(
      relations.items.map(async (relation) => {
        const response = await fetch(`${this.baseUrl}/notes/${relation.note_id}`);
        if (!response.ok) {
          throw new Error("Unable to load note.");
        }
        return (await response.json()) as NoteResponse;
      })
    );

    return notes.map((note) => ({
      id: note.id,
      title: note.title,
      noteType: note.note_type,
      contentMd: note.content_md
    }));
  }

  async listNotesByPlane(planeId: string): Promise<CampaignNoteSummary[]> {
    const relationsResponse = await fetch(`${this.baseUrl}/owners/plane/${planeId}/notes`);
    if (!relationsResponse.ok) {
      throw new Error("Unable to load plane notes.");
    }
    const relations = (await relationsResponse.json()) as ListNoteOwnersResponse;
    if (!relations.items.length) {
      return [];
    }

    const notes = await Promise.all(
      relations.items.map(async (relation) => {
        const response = await fetch(`${this.baseUrl}/notes/${relation.note_id}`);
        if (!response.ok) {
          throw new Error("Unable to load note.");
        }
        return (await response.json()) as NoteResponse;
      })
    );

    return notes.map((note) => ({
      id: note.id,
      title: note.title,
      noteType: note.note_type,
      ownerType: "plane",
      ownerId: planeId,
      ownerName: ""
    }));
  }

  async listPlanesByCampaign(campaignId: string): Promise<CampaignPlaneSummary[]> {
    const plane = await this.fetchPlaneForCampaign(campaignId);
    return [
      {
        id: plane.id,
        name: plane.name,
        description: plane.description
      }
    ];
  }

  async listWorldsByPlane(planeId: string): Promise<PlaneWorldSummary[]> {
    const worlds = await this.listWorlds();
    return worlds
      .filter((world) => world.planeId === planeId)
      .map((world) => ({
        id: world.id,
        name: world.name,
        description: world.description
      }));
  }

  async updateCampaignNote(
    noteId: string,
    input: UpdateCampaignNoteInput
  ): Promise<CampaignNoteDetail> {
    const response = await fetch(`${this.baseUrl}/notes/${noteId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: input.title,
        content_md: input.contentMd,
        note_type: input.noteType,
        metadata_json: []
      })
    });
    if (!response.ok) {
      throw new Error("Unable to update note.");
    }
    const item = (await response.json()) as NoteResponse;
    return {
      id: item.id,
      title: item.title,
      contentMd: item.content_md,
      noteType: item.note_type,
      ownerType: "campaign",
      ownerId: "",
      ownerName: ""
    };
  }

  async createNote(input: CreateNoteInput): Promise<CampaignNoteDetail> {
    const createResponse = await fetch(`${this.baseUrl}/notes`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: input.title,
        content_md: input.contentMd,
        note_type: input.noteType,
        metadata_json: []
      })
    });
    if (!createResponse.ok) {
      throw new Error("Unable to create note.");
    }
    const created = (await createResponse.json()) as NoteResponse;

    const attachResponse = await fetch(`${this.baseUrl}/notes/${created.id}/owners`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        owner_type: input.ownerType,
        owner_id: input.ownerId
      })
    });
    if (!attachResponse.ok) {
      throw new Error("Unable to attach note owner.");
    }

    return {
      id: created.id,
      title: created.title,
      contentMd: created.content_md,
      noteType: created.note_type,
      ownerType: input.ownerType,
      ownerId: input.ownerId,
      ownerName: input.ownerName
    };
  }

  private async listWorlds(): Promise<WorldViewModel[]> {
    const response = await fetch(`${this.baseUrl}/worlds`);
    if (!response.ok) {
      throw new Error("Unable to load worlds.");
    }
    const payload = (await response.json()) as ListWorldsResponse;
    return payload.items.map((item) => ({
      id: item.id,
      planeId: item.plane_id,
      name: item.name,
      description: item.description,
      status: item.status as WorldViewModel["status"],
      version: item.version
    }));
  }

  private async fetchPlaneForCampaign(campaignId: string): Promise<PlaneResponse> {
    const campaignResponse = await fetch(`${this.baseUrl}/campaigns/${campaignId}`);
    if (!campaignResponse.ok) {
      throw new Error("Unable to load campaign.");
    }
    const campaign = (await campaignResponse.json()) as CampaignResponse;

    const worldResponse = await fetch(`${this.baseUrl}/worlds/${campaign.world_id}`);
    if (!worldResponse.ok) {
      throw new Error("Unable to load world.");
    }
    const world = (await worldResponse.json()) as WorldResponse;

    const planeResponse = await fetch(`${this.baseUrl}/planes/${world.plane_id}`);
    if (!planeResponse.ok) {
      throw new Error("Unable to load plane.");
    }
    return (await planeResponse.json()) as PlaneResponse;
  }

  private async resolveWorldIDForPlane(planeId: string): Promise<string> {
    const worlds = await this.listWorlds();
    const world = worlds.find((item) => item.planeId === planeId);
    if (!world) {
      throw new Error("Cannot create campaign without a world associated to plane.");
    }
    return world.id;
  }
}
