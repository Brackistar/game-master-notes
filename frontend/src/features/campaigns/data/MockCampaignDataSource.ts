import type {
  CampaignNoteDetail,
  CampaignNoteSummary,
  CampaignPlaneNote,
  CampaignPlaneSummary,
  CampaignPlayerDetail,
  CampaignPlayerSummary,
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
import type { CampaignDataSource } from "./CampaignDataSource";

const defaultPlaneId = "01HQ8J5QK8W8S6W2A3E0P8P001";

const seedCampaigns: CampaignViewModel[] = [
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A001",
    planeId: defaultPlaneId,
    name: "Ashes of the Ivory Coast",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A002",
    planeId: defaultPlaneId,
    name: "The Glass Crown Conspiracy",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A003",
    planeId: defaultPlaneId,
    name: "Shadows Beneath Iron Harbor",
    version: 1
  },
  {
    id: "01HQ8J5QK8W8S6W2A3E0P8A004",
    planeId: defaultPlaneId,
    name: "Stars Over Thornwatch Bastion",
    version: 1
  }
];

export class MockCampaignDataSource implements CampaignDataSource {
  private planes: PlaneViewModel[];
  private worlds: WorldViewModel[];
  private campaigns: CampaignViewModel[];
  private planeByCampaignId: Map<string, CampaignPlaneNote[]>;
  private notesByCampaignId: Map<string, CampaignNoteDetail[]>;
  private playersByCampaignId: Map<string, CampaignPlayerSummary[]>;
  private planesByCampaignId: Map<string, CampaignPlaneSummary[]>;
  private notesByPlayerId: Map<string, PlayerNoteCard[]>;
  private notesByPlaneId: Map<string, CampaignNoteDetail[]>;

  constructor(initial: CampaignViewModel[] = seedCampaigns) {
    this.planes = [
      {
        id: defaultPlaneId,
        name: "Prime Material Reach",
        description: "Main mortal reality of ports, crowns, and ruined roads.",
        version: 1
      },
      {
        id: "01HQ8J5QK8W8S6W2A3E0P8P002",
        name: "Umbral Verge",
        description: "A shadow-thin mirror realm and omen-laden pathways.",
        version: 1
      }
    ];
    this.worlds = [
      {
        id: "01HQ8J5QK8W8S6W2A3E0P8W001",
        planeId: defaultPlaneId,
        name: "Ivory Storm Coast",
        description: "Storm-ravaged guild coast and relic routes.",
        status: "active",
        version: 1
      },
      {
        id: "01HQ8J5QK8W8S6W2A3E0P8W002",
        planeId: defaultPlaneId,
        name: "Emberglass Dominion",
        description: "Courtly intrigue among prophetic dynasties.",
        status: "active",
        version: 1
      },
      {
        id: "01HQ8J5QK8W8S6W2A3E0P8W003",
        planeId: "01HQ8J5QK8W8S6W2A3E0P8P002",
        name: "Night Reflection",
        description: "A dim mirror world of echoes and fragmented oaths.",
        status: "draft",
        version: 1
      }
    ];
    this.campaigns = initial.map((item) => ({ ...item }));
    this.planeByCampaignId = new Map(
      this.campaigns.map((campaign) => {
        const plane = this.planes.find((item) => item.id === campaign.planeId);
        return [
          campaign.id,
          [
            {
              id: `PLANE-NOTE-${campaign.id}`,
              planeId: campaign.planeId,
              title: plane?.name ?? "Unknown Plane",
              contentMd: plane?.description ?? "No plane description yet."
            }
          ]
        ];
      })
    );
    this.notesByCampaignId = new Map([
      [
        "01HQ8J5QK8W8S6W2A3E0P8A001",
        [
          {
            id: "NOTE-ASH-001",
            title: "Session Zero - Factions",
            noteType: "general",
            contentMd:
              "# Factions\n- Salt Cartel\n- Pilgrim Lanterns\n- Black Oath Navy",
            ownerType: "campaign",
            ownerId: "01HQ8J5QK8W8S6W2A3E0P8A001",
            ownerName: "Ashes of the Ivory Coast"
          }
        ]
      ]
    ]);
    this.playersByCampaignId = new Map([
      [
        "01HQ8J5QK8W8S6W2A3E0P8A001",
        [
          { id: "PLAYER-001", name: "Elara Stormborne" },
          { id: "PLAYER-002", name: "Dain Blackreef" }
        ]
      ]
    ]);
    this.planesByCampaignId = new Map(
      this.campaigns.map((campaign) => {
        const plane = this.planes.find((item) => item.id === campaign.planeId);
        return [
          campaign.id,
          plane
            ? [
                {
                  id: plane.id,
                  name: plane.name,
                  description: plane.description
                }
              ]
            : []
        ];
      })
    );
    this.notesByPlayerId = new Map([
      [
        "PLAYER-001",
        buildPlayerNotes("Elara", 42, "Ashes of the Ivory Coast", "general")
      ]
    ]);
    this.notesByPlaneId = new Map([
      [
        defaultPlaneId,
        [
          {
            id: "NOTE-PLANE-001",
            title: "Prime Reach Atlas",
            noteType: "map",
            contentMd: "# Prime Reach Atlas\nCoastal routes and leyline ports.",
            ownerType: "plane",
            ownerId: defaultPlaneId,
            ownerName: "Prime Material Reach"
          }
        ]
      ]
    ]);
  }

  async listPlanes(): Promise<PlaneViewModel[]> {
    return this.planes.map((item) => ({ ...item }));
  }

  async createPlane(input: CreatePlaneInput): Promise<PlaneViewModel> {
    const created: PlaneViewModel = {
      id: this.generateID(),
      name: input.name.trim(),
      description: input.description.trim(),
      version: 1
    };
    this.planes = [created, ...this.planes];
    return { ...created };
  }

  async updatePlane(input: UpdatePlaneInput): Promise<PlaneViewModel> {
    const index = this.planes.findIndex((item) => item.id === input.id);
    if (index < 0) {
      throw new Error("Plane not found.");
    }
    const updated: PlaneViewModel = {
      ...this.planes[index],
      name: input.name.trim(),
      description: input.description.trim(),
      version: this.planes[index].version + 1
    };
    const next = [...this.planes];
    next[index] = updated;
    this.planes = next;
    return { ...updated };
  }

  async deletePlane(id: string): Promise<void> {
    this.planes = this.planes.filter((item) => item.id !== id);
    this.campaigns = this.campaigns.filter((campaign) => campaign.planeId !== id);
    this.worlds = this.worlds.filter((world) => world.planeId !== id);
  }

  async listCampaigns(): Promise<CampaignViewModel[]> {
    return this.campaigns.map((item) => ({ ...item }));
  }

  async createCampaign(input: CreateCampaignInput): Promise<CampaignViewModel> {
    const created: CampaignViewModel = {
      id: this.generateID(),
      planeId: input.planeId ?? defaultPlaneId,
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

  async listPlaneNotesByCampaign(campaignId: string): Promise<CampaignPlaneNote[]> {
    return (this.planeByCampaignId.get(campaignId) ?? []).map((item) => ({ ...item }));
  }

  async listNotesByCampaign(campaignId: string): Promise<CampaignNoteSummary[]> {
    return (this.notesByCampaignId.get(campaignId) ?? []).map((item) => ({
      id: item.id,
      title: item.title,
      noteType: item.noteType,
      ownerType: item.ownerType,
      ownerId: item.ownerId,
      ownerName: item.ownerName
    }));
  }

  async getNoteById(noteId: string): Promise<CampaignNoteDetail> {
    for (const notes of this.notesByCampaignId.values()) {
      const found = notes.find((item) => item.id === noteId);
      if (found) {
        return { ...found };
      }
    }
    for (const notes of this.notesByPlaneId.values()) {
      const found = notes.find((item) => item.id === noteId);
      if (found) {
        return { ...found };
      }
    }
    for (const [playerId, notes] of this.notesByPlayerId.entries()) {
      const found = notes.find((item) => item.id === noteId);
      if (found) {
        return {
          id: found.id,
          title: found.title,
          contentMd: found.contentMd,
          noteType: found.noteType,
          ownerType: "player",
          ownerId: playerId,
          ownerName: this.getPlayerName(playerId)
        };
      }
    }
    throw new Error("Note not found.");
  }

  async listPlayersByCampaign(campaignId: string): Promise<CampaignPlayerSummary[]> {
    return (this.playersByCampaignId.get(campaignId) ?? []).map((item) => ({ ...item }));
  }

  async getPlayerById(playerId: string): Promise<CampaignPlayerDetail> {
    for (const players of this.playersByCampaignId.values()) {
      const found = players.find((item) => item.id === playerId);
      if (found) {
        return { ...found };
      }
    }
    throw new Error("Player not found.");
  }

  async listNotesByPlayer(playerId: string): Promise<PlayerNoteCard[]> {
    return (this.notesByPlayerId.get(playerId) ?? []).map((item) => ({ ...item }));
  }

  async listNotesByPlane(planeId: string): Promise<CampaignNoteSummary[]> {
    return (this.notesByPlaneId.get(planeId) ?? []).map((item) => ({
      id: item.id,
      title: item.title,
      noteType: item.noteType,
      ownerType: "plane",
      ownerId: item.ownerId,
      ownerName: item.ownerName
    }));
  }

  async listPlanesByCampaign(campaignId: string): Promise<CampaignPlaneSummary[]> {
    return (this.planesByCampaignId.get(campaignId) ?? []).map((item) => ({ ...item }));
  }

  async listWorldsByPlane(planeId: string): Promise<PlaneWorldSummary[]> {
    return this.worlds
      .filter((world) => world.planeId === planeId)
      .map((world) => ({
        id: world.id,
        name: world.name,
        description: world.description
      }));
  }

  async updateCampaignNote(
    noteID: string,
    input: UpdateCampaignNoteInput
  ): Promise<CampaignNoteDetail> {
    for (const [campaignID, notes] of this.notesByCampaignId.entries()) {
      const index = notes.findIndex((item) => item.id === noteID);
      if (index < 0) {
        continue;
      }
      const updated: CampaignNoteDetail = {
        ...notes[index],
        title: input.title,
        contentMd: input.contentMd,
        noteType: input.noteType
      };
      const next = [...notes];
      next[index] = updated;
      this.notesByCampaignId.set(campaignID, next);
      return { ...updated };
    }
    for (const [planeID, notes] of this.notesByPlaneId.entries()) {
      const index = notes.findIndex((item) => item.id === noteID);
      if (index < 0) {
        continue;
      }
      const updated: CampaignNoteDetail = {
        ...notes[index],
        title: input.title,
        contentMd: input.contentMd,
        noteType: input.noteType
      };
      const next = [...notes];
      next[index] = updated;
      this.notesByPlaneId.set(planeID, next);
      return { ...updated };
    }
    throw new Error("Note not found.");
  }

  async createNote(input: CreateNoteInput): Promise<CampaignNoteDetail> {
    const id = this.generateID();
    const created: CampaignNoteDetail = {
      id,
      title: input.title,
      contentMd: input.contentMd,
      noteType: input.noteType,
      ownerType: input.ownerType,
      ownerId: input.ownerId,
      ownerName: input.ownerName
    };

    if (input.ownerType === "campaign") {
      const notes = this.notesByCampaignId.get(input.ownerId) ?? [];
      this.notesByCampaignId.set(input.ownerId, [created, ...notes]);
      return created;
    }
    if (input.ownerType === "plane") {
      const notes = this.notesByPlaneId.get(input.ownerId) ?? [];
      this.notesByPlaneId.set(input.ownerId, [created, ...notes]);
      return created;
    }

    const cards = this.notesByPlayerId.get(input.ownerId) ?? [];
    this.notesByPlayerId.set(input.ownerId, [
      { id, title: input.title, noteType: input.noteType, contentMd: input.contentMd },
      ...cards
    ]);
    return created;
  }

  private getPlayerName(playerID: string): string {
    for (const players of this.playersByCampaignId.values()) {
      const found = players.find((item) => item.id === playerID);
      if (found) {
        return found.name;
      }
    }
    return "Unknown Player";
  }

  private generateID(): string {
    const rand = Math.random().toString(36).slice(2, 12).toUpperCase();
    const ts = Date.now().toString(36).toUpperCase();
    return `MOCK${ts}${rand}`.slice(0, 26);
  }
}

function buildPlayerNotes(
  playerName: string,
  count: number,
  campaignName: string,
  noteType: string
): PlayerNoteCard[] {
  return Array.from({ length: count }).map((_, index) => ({
    id: `${playerName.toUpperCase()}-NOTE-${String(index + 1).padStart(3, "0")}`,
    title: `${playerName} Chronicle ${index + 1}`,
    noteType,
    contentMd:
      `# ${playerName} Chronicle ${index + 1}\n` +
      `Campaign: ${campaignName}\n\n` +
      `- Current objective thread ${index + 1}\n` +
      `- Relationship hooks and unresolved tensions\n` +
      `- Inventory clues and faction leverage points`
  }));
}
