import { useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import { CreateCampaignForm } from "../../features/campaigns/CreateCampaignForm";
import { noteTypeOptions } from "../../features/campaigns/noteTypes";
import { PlayerNotesGrid } from "../../features/players/PlayerNotesGrid";
import type {
  CampaignNoteDetail,
  CampaignPlaneNote,
  CampaignPlayerDetail,
  CreateNoteInput,
  CreatePlaneInput,
  NoteOwnerType,
  PlaneViewModel,
  PlaneWorldSummary,
  PlayerNoteCard,
  UpdateCampaignNoteInput,
  UpdatePlaneInput
} from "../../models/communication/campaign_models";
import styles from "./MainPanel.module.css";

type MainPanelProps = {
  mode: "empty" | "create-campaign" | "create-note" | "create-plane";
  onCreateCampaign: (name: string) => Promise<void>;
  onCancelCreateCampaign: () => void;
  onCancelCreateNote: () => void;
  onCancelCreatePlane: () => void;
  onCreateNote: (input: CreateNoteInput) => Promise<void>;
  onCreatePlane: (input: CreatePlaneInput) => Promise<void>;
  onUpdatePlane: (input: UpdatePlaneInput) => Promise<void>;
  onDeletePlane: (planeId: string) => Promise<void>;
  createNoteTargets: {
    campaigns: { id: string; name: string }[];
    players: { id: string; name: string }[];
    planes: { id: string; name: string }[];
  };
  selectedPlane: PlaneViewModel | null;
  selectedWorld: PlaneWorldSummary | null;
  selectedCampaignName: string | null;
  planeNotes: CampaignPlaneNote[];
  worldsBySelectedPlane: PlaneWorldSummary[];
  onSelectWorld: (worldID: string) => void;
  onBackToPlaneFromWorld: () => void;
  selectedPlayer: CampaignPlayerDetail | null;
  playerNotes: PlayerNoteCard[];
  playerNotesLoading: boolean;
  activeNote: CampaignNoteDetail | null;
  activeNoteSource: "campaign" | "plane" | "player" | null;
  onSelectPlaneNote: (planeNote: CampaignPlaneNote) => void;
  onBackToPlaneList: () => void;
  onSelectPlayerNote: (noteId: string) => void;
  onSaveCampaignNote: (
    noteId: string,
    input: UpdateCampaignNoteInput
  ) => Promise<void>;
};

export function MainPanel(props: MainPanelProps) {
  const [isEditingNote, setIsEditingNote] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [editContent, setEditContent] = useState("");
  const [editNoteType, setEditNoteType] =
    useState<(typeof noteTypeOptions)[number]>("general");
  const [editError, setEditError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [createTitle, setCreateTitle] = useState("");
  const [createContent, setCreateContent] = useState("");
  const [createType, setCreateType] = useState<(typeof noteTypeOptions)[number]>("general");
  const [createOwnerType, setCreateOwnerType] = useState<NoteOwnerType>("campaign");
  const [createOwnerID, setCreateOwnerID] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);

  const [createPlaneName, setCreatePlaneName] = useState("");
  const [createPlaneDescription, setCreatePlaneDescription] = useState("");
  const [createPlaneError, setCreatePlaneError] = useState<string | null>(null);
  const [editPlaneName, setEditPlaneName] = useState("");
  const [editPlaneDescription, setEditPlaneDescription] = useState("");
  const [planeError, setPlaneError] = useState<string | null>(null);

  useEffect(() => {
    if (!props.activeNote) {
      setIsEditingNote(false);
      setEditTitle("");
      setEditContent("");
      setEditNoteType("general");
      setEditError(null);
      setSaving(false);
      return;
    }
    setEditTitle(props.activeNote.title);
    setEditContent(props.activeNote.contentMd);
    setEditNoteType(
      (props.activeNote.noteType as (typeof noteTypeOptions)[number]) ?? "general"
    );
    setEditError(null);
    setSaving(false);
  }, [props.activeNote]);

  useEffect(() => {
    if (!props.selectedPlane) {
      setEditPlaneName("");
      setEditPlaneDescription("");
      setPlaneError(null);
      return;
    }
    setEditPlaneName(props.selectedPlane.name);
    setEditPlaneDescription(props.selectedPlane.description);
    setPlaneError(null);
  }, [props.selectedPlane]);

  if (props.mode === "create-campaign") {
    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <h2>Create Campaign</h2>
          <p>Set campaign name to add it into the left panel workflow.</p>
          <CreateCampaignForm
            onCreate={props.onCreateCampaign}
            onCancel={props.onCancelCreateCampaign}
          />
        </div>
      </main>
    );
  }

  if (props.mode === "create-note") {
    const ownerTargets =
      createOwnerType === "campaign"
        ? props.createNoteTargets.campaigns
        : createOwnerType === "player"
          ? props.createNoteTargets.players
          : props.createNoteTargets.planes;

    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <h2>Create Note</h2>
          <p>Create a note and attach it to one target entity.</p>
          <form
            className={styles.editForm}
            onSubmit={async (event) => {
              event.preventDefault();
              const title = createTitle.trim();
              const contentMd = createContent.trim();
              if (title.length < 3 || title.length > 50) {
                setCreateError("Title must be between 3 and 50 characters.");
                return;
              }
              if (!contentMd) {
                setCreateError("Note content is required.");
                return;
              }
              if (!createOwnerID) {
                setCreateError("Select a target entity.");
                return;
              }
              const ownerName =
                ownerTargets.find((item) => item.id === createOwnerID)?.name ?? "";
              try {
                await props.onCreateNote({
                  title,
                  contentMd,
                  noteType: createType,
                  ownerType: createOwnerType,
                  ownerId: createOwnerID,
                  ownerName
                });
                setCreateTitle("");
                setCreateContent("");
                setCreateType("general");
                setCreateError(null);
              } catch {
                setCreateError("Unable to create note right now.");
              }
            }}
          >
            <label className={styles.formLabel}>Title</label>
            <input
              value={createTitle}
              onChange={(event) => setCreateTitle(event.target.value)}
              maxLength={50}
            />
            <label className={styles.formLabel}>Note Type</label>
            <select
              value={createType}
              onChange={(event) =>
                setCreateType(event.target.value as (typeof noteTypeOptions)[number])
              }
              className={styles.selectInput}
            >
              {noteTypeOptions.map((item) => (
                <option key={item} value={item}>
                  {item}
                </option>
              ))}
            </select>
            <label className={styles.formLabel}>Attach To</label>
            <select
              value={createOwnerType}
              onChange={(event) => {
                setCreateOwnerType(event.target.value as NoteOwnerType);
                setCreateOwnerID("");
              }}
              className={styles.selectInput}
            >
              <option value="campaign">Campaign</option>
              <option value="player">Player</option>
              <option value="plane">Plane</option>
            </select>
            <label className={styles.formLabel}>Target Entity</label>
            <select
              value={createOwnerID}
              onChange={(event) => setCreateOwnerID(event.target.value)}
              className={styles.selectInput}
            >
              <option value="">Select target</option>
              {ownerTargets.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
            <label className={styles.formLabel}>Markdown Content</label>
            <textarea
              className={styles.editTextarea}
              value={createContent}
              onChange={(event) => setCreateContent(event.target.value)}
              rows={10}
            />
            {createError ? <p className={styles.formError}>{createError}</p> : null}
            <div className={styles.formActions}>
              <button type="button" onClick={props.onCancelCreateNote}>
                Cancel
              </button>
              <button type="submit">Create</button>
            </div>
          </form>
        </div>
      </main>
    );
  }

  if (props.mode === "create-plane") {
    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <h2>Create Plane</h2>
          <form
            className={styles.editForm}
            onSubmit={async (event) => {
              event.preventDefault();
              const name = createPlaneName.trim();
              if (name.length < 3 || name.length > 50) {
                setCreatePlaneError("Plane name must be between 3 and 50 characters.");
                return;
              }
              try {
                await props.onCreatePlane({
                  name,
                  description: createPlaneDescription.trim()
                });
                setCreatePlaneName("");
                setCreatePlaneDescription("");
                setCreatePlaneError(null);
              } catch {
                setCreatePlaneError("Unable to create plane right now.");
              }
            }}
          >
            <label className={styles.formLabel}>Name</label>
            <input
              value={createPlaneName}
              onChange={(event) => setCreatePlaneName(event.target.value)}
              maxLength={50}
            />
            <label className={styles.formLabel}>Description</label>
            <textarea
              className={styles.editTextarea}
              value={createPlaneDescription}
              onChange={(event) => setCreatePlaneDescription(event.target.value)}
              rows={8}
            />
            {createPlaneError ? <p className={styles.formError}>{createPlaneError}</p> : null}
            <div className={styles.formActions}>
              <button type="button" onClick={props.onCancelCreatePlane}>
                Cancel
              </button>
              <button type="submit">Create Plane</button>
            </div>
          </form>
        </div>
      </main>
    );
  }

  if (props.selectedPlayer && !props.activeNote) {
    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <section className={styles.playerInfoSection}>
            <h2>{props.selectedPlayer.name}</h2>
            <p>Player ID: {props.selectedPlayer.id}</p>
            <p>Associated notes: {props.playerNotes.length}</p>
          </section>
          <section className={styles.playerNotesSection}>
            <div className={styles.playerNotesHeader}>Player Notes</div>
            {props.playerNotesLoading ? <p>Loading player notes...</p> : null}
            {!props.playerNotesLoading && props.playerNotes.length === 0 ? (
              <p>No notes associated with this player yet.</p>
            ) : null}
            {!props.playerNotesLoading && props.playerNotes.length > 0 ? (
              <PlayerNotesGrid
                notes={props.playerNotes}
                onSelectNote={props.onSelectPlayerNote}
              />
            ) : null}
          </section>
        </div>
      </main>
    );
  }

  if (props.activeNote) {
    const activeNote = props.activeNote;

    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <div className={styles.noteHeader}>
            <h2>{activeNote.title}</h2>
            <div className={styles.noteHeaderActions}>
              <button
                type="button"
                onClick={() => setIsEditingNote((current) => !current)}
              >
                {isEditingNote ? "Preview" : "Edit"}
              </button>
              <button type="button" onClick={props.onBackToPlaneList}>
                Back
              </button>
            </div>
          </div>
          <div className={styles.ownerChipRow}>
            <span className={styles.ownerChip}>
              Attached To: {activeNote.ownerType} - {activeNote.ownerName}
            </span>
          </div>
          <div className={styles.noteMeta}>Type: {activeNote.noteType}</div>
          {isEditingNote ? (
            <form
              className={styles.editForm}
              onSubmit={async (event) => {
                event.preventDefault();
                const title = editTitle.trim();
                const contentMd = editContent.trim();
                if (title.length < 3 || title.length > 50) {
                  setEditError("Title must be between 3 and 50 characters.");
                  return;
                }
                if (contentMd.length === 0) {
                  setEditError("Note content is required.");
                  return;
                }
                try {
                  setSaving(true);
                  setEditError(null);
                  await props.onSaveCampaignNote(activeNote.id, {
                    title,
                    contentMd,
                    noteType: editNoteType
                  });
                  setIsEditingNote(false);
                } catch {
                  setEditError("Unable to save note changes right now.");
                } finally {
                  setSaving(false);
                }
              }}
            >
              <label htmlFor="note-title" className={styles.formLabel}>
                Title
              </label>
              <input
                id="note-title"
                value={editTitle}
                onChange={(event) => setEditTitle(event.target.value)}
                maxLength={50}
              />
              <label htmlFor="note-content" className={styles.formLabel}>
                Markdown Content
              </label>
              <label htmlFor="note-type" className={styles.formLabel}>
                Note Type
              </label>
              <select
                id="note-type"
                className={styles.selectInput}
                value={editNoteType}
                onChange={(event) =>
                  setEditNoteType(event.target.value as (typeof noteTypeOptions)[number])
                }
              >
                {noteTypeOptions.map((item) => (
                  <option key={item} value={item}>
                    {item}
                  </option>
                ))}
              </select>
              <textarea
                id="note-content"
                className={styles.editTextarea}
                value={editContent}
                onChange={(event) => setEditContent(event.target.value)}
                rows={12}
              />
              {editError ? <p className={styles.formError}>{editError}</p> : null}
              <div className={styles.formActions}>
                <button
                  type="button"
                  onClick={() => {
                    setEditTitle(activeNote.title);
                    setEditContent(activeNote.contentMd);
                    setEditNoteType(
                      (activeNote.noteType as (typeof noteTypeOptions)[number]) ??
                        "general"
                    );
                    setEditError(null);
                    setIsEditingNote(false);
                  }}
                  disabled={saving}
                >
                  Cancel
                </button>
                <button type="submit" disabled={saving}>
                  {saving ? "Saving..." : "Save"}
                </button>
              </div>
            </form>
          ) : (
            <article className={styles.markdownContent}>
              <ReactMarkdown skipHtml>{activeNote.contentMd}</ReactMarkdown>
            </article>
          )}
        </div>
      </main>
    );
  }

  if (props.selectedCampaignName) {
    if (props.selectedWorld) {
      const selectedWorld = props.selectedWorld;
      return (
        <main className={styles.mainPanel}>
          <div className={styles.centerCard}>
            <div className={styles.noteHeader}>
              <h2>{selectedWorld.name}</h2>
              <div className={styles.noteHeaderActions}>
                <button type="button" onClick={props.onBackToPlaneFromWorld}>
                  Back To Plane
                </button>
              </div>
            </div>
            <section className={styles.playerInfoSection}>
              <p>World ID: {selectedWorld.id}</p>
              <p>{selectedWorld.description || "No world description yet."}</p>
            </section>
            <section className={styles.playerNotesSection}>
              <div className={styles.playerNotesHeader}>World Workspace</div>
              <p>World notes can now route to this main world view for editing workflows.</p>
            </section>
          </div>
        </main>
      );
    }

    if (props.selectedPlane) {
      const selectedPlane = props.selectedPlane;
      return (
        <main className={styles.mainPanel}>
          <div className={styles.centerCard}>
            <section className={styles.playerInfoSection}>
              <h2>{selectedPlane.name}</h2>
              <p>Plane ID: {selectedPlane.id}</p>
              <p>{selectedPlane.description || "No plane description yet."}</p>
            </section>
            <section className={styles.playerNotesSection}>
              <div className={styles.playerNotesHeader}>Worlds In This Plane</div>
              {props.worldsBySelectedPlane.length ? (
                <ul className={styles.worldNoteList}>
                  {props.worldsBySelectedPlane.map((world) => (
                    <li key={world.id}>
                      <button
                        type="button"
                        className={styles.worldNoteButton}
                        onClick={() => props.onSelectWorld(world.id)}
                      >
                        <strong>{world.name}</strong>
                        {world.description ? <div>{world.description}</div> : null}
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p>No worlds associated with this plane yet.</p>
              )}
            </section>
            <form
              className={styles.editForm}
              onSubmit={async (event) => {
                event.preventDefault();
                const name = editPlaneName.trim();
                if (name.length < 3 || name.length > 50) {
                  setPlaneError("Plane name must be between 3 and 50 characters.");
                  return;
                }
                try {
                  await props.onUpdatePlane({
                    id: selectedPlane.id,
                    name,
                    description: editPlaneDescription.trim()
                  });
                  setPlaneError(null);
                } catch {
                  setPlaneError("Unable to update plane right now.");
                }
              }}
            >
              <label className={styles.formLabel}>Plane Name</label>
              <input
                value={editPlaneName}
                onChange={(event) => setEditPlaneName(event.target.value)}
                maxLength={50}
              />
              <label className={styles.formLabel}>Plane Description</label>
              <textarea
                className={styles.editTextarea}
                value={editPlaneDescription}
                onChange={(event) => setEditPlaneDescription(event.target.value)}
                rows={5}
              />
              {planeError ? <p className={styles.formError}>{planeError}</p> : null}
              <div className={styles.formActions}>
                <button
                  type="button"
                  onClick={() => {
                    void props.onDeletePlane(selectedPlane.id);
                  }}
                >
                  Delete Plane
                </button>
                <button type="submit">Save Plane</button>
              </div>
            </form>
          </div>
        </main>
      );
    }

    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <h2>{props.selectedCampaignName}</h2>
          <p>Plane notes associated with this campaign.</p>
          {props.planeNotes.length ? (
            <ul className={styles.worldNoteList}>
              {props.planeNotes.map((planeNote) => (
                <li key={planeNote.id}>
                  <button
                    type="button"
                    className={styles.worldNoteButton}
                    onClick={() => props.onSelectPlaneNote(planeNote)}
                  >
                    {planeNote.title}
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <p>No plane notes available for this campaign yet.</p>
          )}
        </div>
      </main>
    );
  }

  return (
    <main className={styles.mainPanel}>
      <div className={styles.centerCard}>
        <h2>Campaign Notes Workspace</h2>
        <p>Select a campaign on the left or click Add to start a new campaign.</p>
      </div>
    </main>
  );
}
