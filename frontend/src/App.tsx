import { useCallback, useEffect, useMemo, useState } from "react";
import { CampaignPanel } from "./features/campaigns/CampaignPanel";
import { createCampaignDataSource } from "./features/campaigns/data/createCampaignDataSource";
import type {
  CampaignNoteDetail,
  CampaignNoteSummary,
  CampaignPlaneNote,
  CampaignPlaneSummary,
  CampaignPlayerDetail,
  CampaignPlayerSummary,
  CampaignViewModel,
  CreateNoteInput,
  CreatePlaneInput,
  NoteOwnerType,
  PlaneViewModel,
  PlaneWorldSummary,
  PlayerNoteCard,
  UpdateCampaignNoteInput,
  UpdatePlaneInput
} from "./models/communication/campaign_models";
import { BottomBar, type CampaignContextView } from "./ui/layout/BottomBar";
import { MainPanel } from "./ui/layout/MainPanel";
import { RightPanel } from "./ui/layout/RightPanel";
import styles from "./ui/layout/AppShell.module.css";

type CenterMode = "empty" | "create-campaign" | "create-note" | "create-plane";
type NoteSource = "campaign" | "player" | "plane";

export function App() {
  const campaignDataSource = useMemo(() => createCampaignDataSource(), []);

  const [campaigns, setCampaigns] = useState<CampaignViewModel[]>([]);
  const [selectedCampaignID, setSelectedCampaignID] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [centerMode, setCenterMode] = useState<CenterMode>("empty");
  const [loadError, setLoadError] = useState<string | null>(null);

  const [planeNotes, setPlaneNotes] = useState<CampaignPlaneNote[]>([]);
  const [campaignNotes, setCampaignNotes] = useState<CampaignNoteSummary[]>([]);
  const [campaignPlayers, setCampaignPlayers] = useState<CampaignPlayerSummary[]>([]);
  const [campaignPlanes, setCampaignPlanes] = useState<CampaignPlaneSummary[]>([]);
  const [selectedPlaneID, setSelectedPlaneID] = useState<string | null>(null);
  const [worldsBySelectedPlane, setWorldsBySelectedPlane] = useState<PlaneWorldSummary[]>([]);
  const [selectedWorldID, setSelectedWorldID] = useState<string | null>(null);

  const [activeNote, setActiveNote] = useState<CampaignNoteDetail | null>(null);
  const [activeNoteSource, setActiveNoteSource] = useState<NoteSource | null>(null);
  const [selectedNoteID, setSelectedNoteID] = useState<string | null>(null);
  const [workspaceLoading, setWorkspaceLoading] = useState(false);

  const [rightPanelContext, setRightPanelContext] =
    useState<CampaignContextView>("notes");
  const [selectedPlayerID, setSelectedPlayerID] = useState<string | null>(null);
  const [selectedPlayer, setSelectedPlayer] = useState<CampaignPlayerDetail | null>(null);
  const [playerNotes, setPlayerNotes] = useState<PlayerNoteCard[]>([]);
  const [playerNotesLoading, setPlayerNotesLoading] = useState(false);
  const [allPlanes, setAllPlanes] = useState<PlaneViewModel[]>([]);

  const selectedCampaign = useMemo(
    () => campaigns.find((campaign) => campaign.id === selectedCampaignID) ?? null,
    [campaigns, selectedCampaignID]
  );

  const selectedPlane = useMemo(
    () => allPlanes.find((plane) => plane.id === selectedPlaneID) ?? null,
    [allPlanes, selectedPlaneID]
  );

  const selectedWorld = useMemo(
    () => worldsBySelectedPlane.find((world) => world.id === selectedWorldID) ?? null,
    [selectedWorldID, worldsBySelectedPlane]
  );

  const selectedCampaignExists = useMemo(
    () =>
      selectedCampaignID !== null &&
      campaigns.some((campaign) => campaign.id === selectedCampaignID),
    [campaigns, selectedCampaignID]
  );

  const refreshCampaignContext = useCallback(
    async (campaignID: string) => {
      const [nextPlaneNotes, nextCampaignNotes, nextPlayers, nextPlanes] =
        await Promise.all([
          campaignDataSource.listPlaneNotesByCampaign(campaignID),
          campaignDataSource.listNotesByCampaign(campaignID),
          campaignDataSource.listPlayersByCampaign(campaignID),
          campaignDataSource.listPlanesByCampaign(campaignID)
        ]);
      setPlaneNotes(nextPlaneNotes);
      setCampaignNotes(nextCampaignNotes);
      setCampaignPlayers(nextPlayers);
      setCampaignPlanes(nextPlanes);
      setSelectedPlaneID(nextPlanes[0]?.id ?? null);
    },
    [campaignDataSource]
  );

  const refreshPlayerNotes = useCallback(
    async (playerID: string) => {
      const notes = await campaignDataSource.listNotesByPlayer(playerID);
      setPlayerNotes(notes);
    },
    [campaignDataSource]
  );

  useEffect(() => {
    void (async () => {
      try {
        setLoadError(null);
        const [items, planes] = await Promise.all([
          campaignDataSource.listCampaigns(),
          campaignDataSource.listPlanes()
        ]);
        setCampaigns(items);
        setAllPlanes(planes);
        setSelectedCampaignID(items[0]?.id ?? null);
      } catch {
        setLoadError("Unable to load campaigns.");
      }
    })();
  }, [campaignDataSource]);

  useEffect(() => {
    if (!selectedCampaignID) {
      setPlaneNotes([]);
      setCampaignNotes([]);
      setCampaignPlayers([]);
      setCampaignPlanes([]);
      setSelectedPlaneID(null);
      setWorldsBySelectedPlane([]);
      setSelectedWorldID(null);
      setSelectedPlayerID(null);
      setSelectedPlayer(null);
      setPlayerNotes([]);
      setSelectedNoteID(null);
      setActiveNote(null);
      setActiveNoteSource(null);
      return;
    }

    void (async () => {
      try {
        setWorkspaceLoading(true);
        await refreshCampaignContext(selectedCampaignID);
        setSelectedPlayerID(null);
        setSelectedPlayer(null);
        setPlayerNotes([]);
        setSelectedNoteID(null);
        setActiveNote(null);
        setActiveNoteSource(null);
      } catch {
        setLoadError("Unable to load campaign workspace.");
      } finally {
        setWorkspaceLoading(false);
      }
    })();
  }, [refreshCampaignContext, selectedCampaignID]);

  useEffect(() => {
    if (!selectedPlaneID) {
      setWorldsBySelectedPlane([]);
      setSelectedWorldID(null);
      return;
    }
    void (async () => {
      try {
        const worlds = await campaignDataSource.listWorldsByPlane(selectedPlaneID);
        setWorldsBySelectedPlane(worlds);
        setSelectedWorldID(null);
      } catch {
        setLoadError("Unable to load plane worlds.");
      }
    })();
  }, [campaignDataSource, selectedPlaneID]);

  const handleReorderCampaigns = useCallback(
    async (idsInOrder: string[]) => {
      const byID = new Map(campaigns.map((item) => [item.id, item]));
      const sorted = idsInOrder
        .map((id) => byID.get(id))
        .filter((item): item is CampaignViewModel => item !== undefined);
      if (sorted.length !== campaigns.length) {
        return;
      }
      setCampaigns(sorted);
      try {
        await campaignDataSource.reorderCampaigns(idsInOrder);
      } catch {
        setLoadError("Unable to save campaign order.");
      }
    },
    [campaignDataSource, campaigns]
  );

  const handleCreateCampaign = useCallback(
    async (name: string) => {
      const created = await campaignDataSource.createCampaign({
        name,
        planeId: selectedPlaneID ?? selectedCampaign?.planeId
      });
      setCampaigns((previous) => [created, ...previous]);
      setSelectedCampaignID(created.id);
      setCenterMode("empty");
      setSearchQuery("");
    },
    [campaignDataSource, selectedCampaign?.planeId, selectedPlaneID]
  );

  const handleDeleteCampaign = useCallback(
    async (id: string) => {
      try {
        await campaignDataSource.deleteCampaign(id);
        setCampaigns((previous) => previous.filter((item) => item.id !== id));
        setSelectedCampaignID((current) => (current === id ? null : current));
      } catch {
        setLoadError("Unable to delete campaign.");
      }
    },
    [campaignDataSource]
  );

  const handleSelectCampaign = useCallback((id: string) => {
    setSelectedCampaignID(id);
    setSelectedPlayerID(null);
    setSelectedPlayer(null);
    setSelectedPlaneID(null);
    setWorldsBySelectedPlane([]);
    setSelectedWorldID(null);
    setPlayerNotes([]);
    setActiveNote(null);
    setActiveNoteSource(null);
    setSelectedNoteID(null);
    setRightPanelContext("notes");
  }, []);

  const handleSelectNote = useCallback(
    async (
      noteID: string,
      source: NoteSource,
      owner: { type: NoteOwnerType; id: string; name: string }
    ) => {
      try {
        const note = await campaignDataSource.getNoteById(noteID);
        setSelectedNoteID(noteID);
        setActiveNote({
          ...note,
          ownerType: owner.type,
          ownerId: owner.id,
          ownerName: owner.name
        });
        if (source === "campaign") {
          setSelectedPlayer(null);
          setSelectedPlayerID(null);
        }
        setActiveNoteSource(source);
      } catch {
        setLoadError("Unable to load note content.");
      }
    },
    [campaignDataSource]
  );

  const handleSelectPlayer = useCallback(
    async (playerID: string) => {
      try {
        setPlayerNotesLoading(true);
        const [player, notes] = await Promise.all([
          campaignDataSource.getPlayerById(playerID),
          campaignDataSource.listNotesByPlayer(playerID)
        ]);
        setSelectedPlayerID(playerID);
        setSelectedPlayer(player);
        setPlayerNotes(notes);
        setSelectedPlaneID(null);
        setWorldsBySelectedPlane([]);
        setSelectedWorldID(null);
        setActiveNote(null);
        setSelectedNoteID(null);
        setActiveNoteSource(null);
      } catch {
        setLoadError("Unable to load player workspace.");
      } finally {
        setPlayerNotesLoading(false);
      }
    },
    [campaignDataSource]
  );

  const handleSaveNote = useCallback(
    async (noteID: string, input: UpdateCampaignNoteInput) => {
      const updated = await campaignDataSource.updateCampaignNote(noteID, input);
      setActiveNote((previous) =>
        previous
          ? {
              ...updated,
              ownerType: previous.ownerType,
              ownerId: previous.ownerId,
              ownerName: previous.ownerName
            }
          : updated
      );

      if (selectedCampaignID) {
        await refreshCampaignContext(selectedCampaignID);
      }
      if (selectedPlayerID) {
        await refreshPlayerNotes(selectedPlayerID);
      }
    },
    [
      campaignDataSource,
      refreshCampaignContext,
      refreshPlayerNotes,
      selectedCampaignID,
      selectedPlayerID
    ]
  );

  const handleSelectPlane = useCallback((planeID: string) => {
    setSelectedPlaneID(planeID);
    setSelectedWorldID(null);
    setSelectedPlayer(null);
    setSelectedPlayerID(null);
    setActiveNote(null);
    setActiveNoteSource(null);
    setSelectedNoteID(null);
  }, []);

  const handleCreatePlane = useCallback(
    async (input: CreatePlaneInput) => {
      const created = await campaignDataSource.createPlane(input);
      setAllPlanes((previous) => [created, ...previous]);
      setCampaignPlanes((previous) => [
        { id: created.id, name: created.name, description: created.description },
        ...previous
      ]);
      setSelectedPlaneID(created.id);
      setCenterMode("empty");
      if (selectedCampaignID) {
        await refreshCampaignContext(selectedCampaignID);
      }
    },
    [campaignDataSource, refreshCampaignContext, selectedCampaignID]
  );

  const handleUpdatePlane = useCallback(
    async (input: UpdatePlaneInput) => {
      const updated = await campaignDataSource.updatePlane(input);
      setAllPlanes((previous) =>
        previous.map((plane) => (plane.id === updated.id ? updated : plane))
      );
      setCampaignPlanes((previous) =>
        previous.map((plane) =>
          plane.id === updated.id
            ? { ...plane, name: updated.name, description: updated.description }
            : plane
        )
      );
    },
    [campaignDataSource]
  );

  const handleDeletePlane = useCallback(
    async (planeID: string) => {
      await campaignDataSource.deletePlane(planeID);
      setAllPlanes((previous) => previous.filter((plane) => plane.id !== planeID));
      setCampaignPlanes((previous) => previous.filter((plane) => plane.id !== planeID));
      setSelectedPlaneID((current) => (current === planeID ? null : current));
      setWorldsBySelectedPlane([]);
      setSelectedWorldID(null);
    },
    [campaignDataSource]
  );

  const handleCreateNote = useCallback(
    async (input: CreateNoteInput) => {
      const created = await campaignDataSource.createNote(input);
      if (selectedCampaignID) {
        await refreshCampaignContext(selectedCampaignID);
      }
      if (input.ownerType === "player") {
        setRightPanelContext("players");
        setSelectedPlayerID(input.ownerId);
        const player = await campaignDataSource.getPlayerById(input.ownerId);
        setSelectedPlayer(player);
        await refreshPlayerNotes(input.ownerId);
      } else if (input.ownerType === "plane") {
        setRightPanelContext("planes");
      } else {
        setRightPanelContext("notes");
      }

      setActiveNote(created);
      setActiveNoteSource(
        input.ownerType === "campaign"
          ? "campaign"
          : input.ownerType === "player"
            ? "player"
            : "plane"
      );
      setSelectedNoteID(created.id);
      setCenterMode("empty");
    },
    [
      campaignDataSource,
      refreshCampaignContext,
      refreshPlayerNotes,
      selectedCampaignID
    ]
  );

  return (
    <div className={styles.appShell}>
      <div className={styles.workspaceRow}>
        <CampaignPanel
          campaigns={campaigns}
          selectedCampaignId={selectedCampaignID}
          onSelectCampaign={handleSelectCampaign}
          onReorderCampaigns={(idsInOrder) => {
            void handleReorderCampaigns(idsInOrder);
          }}
          searchQuery={searchQuery}
          onSearchQueryChange={setSearchQuery}
          onAddCampaign={() => setCenterMode("create-campaign")}
          onDeleteCampaign={(id) => {
            void handleDeleteCampaign(id);
          }}
        />

        <MainPanel
          mode={centerMode}
          onCreateCampaign={handleCreateCampaign}
          onCancelCreateCampaign={() => setCenterMode("empty")}
          onCancelCreateNote={() => setCenterMode("empty")}
          onCancelCreatePlane={() => setCenterMode("empty")}
          onCreateNote={handleCreateNote}
          onCreatePlane={handleCreatePlane}
          onUpdatePlane={handleUpdatePlane}
          onDeletePlane={handleDeletePlane}
          createNoteTargets={{
            campaigns: selectedCampaign
              ? [{ id: selectedCampaign.id, name: selectedCampaign.name }]
              : [],
            players: campaignPlayers.map((player) => ({ id: player.id, name: player.name })),
            planes: campaignPlanes.map((plane) => ({ id: plane.id, name: plane.name }))
          }}
          selectedPlane={selectedPlane}
          selectedWorld={selectedWorld}
          selectedCampaignName={selectedCampaign?.name ?? null}
          planeNotes={planeNotes}
          worldsBySelectedPlane={worldsBySelectedPlane}
          onSelectWorld={(worldID) => {
            setSelectedWorldID(worldID);
          }}
          onBackToPlaneFromWorld={() => {
            setSelectedWorldID(null);
          }}
          selectedPlayer={selectedPlayer}
          playerNotes={playerNotes}
          playerNotesLoading={playerNotesLoading}
          activeNote={activeNote}
          activeNoteSource={activeNoteSource}
          onSelectPlaneNote={(planeNote) => {
            setActiveNote({
              id: planeNote.id,
              title: planeNote.title,
              contentMd: planeNote.contentMd,
              noteType: "plane",
              ownerType: "plane",
              ownerId: planeNote.planeId,
              ownerName: planeNote.title
            });
            setActiveNoteSource("plane");
            setSelectedNoteID(planeNote.id);
          }}
          onBackToPlaneList={() => {
            setActiveNote(null);
            setActiveNoteSource(null);
            setSelectedNoteID(null);
          }}
          onSelectPlayerNote={(noteID) => {
            if (!selectedPlayer) {
              return;
            }
            void handleSelectNote(noteID, "player", {
              type: "player",
              id: selectedPlayer.id,
              name: selectedPlayer.name
            });
          }}
          onSaveCampaignNote={handleSaveNote}
        />

        <RightPanel
          contextView={rightPanelContext}
          selectedCampaignName={selectedCampaign?.name ?? null}
          notes={campaignNotes}
          players={campaignPlayers}
          planes={campaignPlanes}
          selectedNoteId={selectedNoteID}
          selectedPlayerId={selectedPlayerID}
          selectedPlaneId={selectedPlaneID}
          onSelectNote={(noteID) => {
            const owner = campaignNotes.find((item) => item.id === noteID);
            if (!owner) {
              return;
            }
            void handleSelectNote(noteID, "campaign", {
              type: owner.ownerType,
              id: owner.ownerId,
              name: owner.ownerName || selectedCampaign?.name || "Campaign"
            });
          }}
          onSelectPlayer={(playerID) => {
            void handleSelectPlayer(playerID);
          }}
          onSelectPlane={handleSelectPlane}
          onCreatePlane={() => {
            setCenterMode("create-plane");
            setSelectedPlaneID(null);
            setWorldsBySelectedPlane([]);
            setSelectedWorldID(null);
            setActiveNote(null);
            setActiveNoteSource(null);
            setSelectedNoteID(null);
          }}
          loading={workspaceLoading}
        />
      </div>

      {loadError ? <div className={styles.errorBar}>{loadError}</div> : null}
      {!selectedCampaignExists && campaigns.length > 0 ? (
        <div className={styles.errorBar}>Select a campaign to continue.</div>
      ) : null}

      <BottomBar
        activeContextView={rightPanelContext}
        onChangeContextView={setRightPanelContext}
        onQuickAddNote={() => setCenterMode("create-note")}
      />
    </div>
  );
}
