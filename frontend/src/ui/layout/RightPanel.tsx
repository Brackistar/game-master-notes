import type {
  CampaignNoteSummary,
  CampaignPlaneSummary,
  CampaignPlayerSummary,
} from "../../models/communication/campaign_models";
import type { CampaignContextView } from "./BottomBar";
import styles from "./RightPanel.module.css";

type RightPanelProps = {
  contextView: CampaignContextView;
  selectedCampaignName: string | null;
  notes: CampaignNoteSummary[];
  players: CampaignPlayerSummary[];
  planes: CampaignPlaneSummary[];
  selectedNoteId: string | null;
  selectedPlayerId: string | null;
  selectedPlaneId: string | null;
  onSelectNote: (noteId: string) => void;
  onSelectPlayer: (playerId: string) => void;
  onSelectPlane: (planeId: string) => void;
  onCreatePlane: () => void;
  loading: boolean;
};

export function RightPanel(props: RightPanelProps) {
  const title =
    props.contextView === "notes"
      ? "Campaign Notes"
      : props.contextView === "players"
        ? "Campaign Players"
        : "Campaign Planes";

  return (
    <aside className={styles.rightPanel} aria-label="Context panel">
      <h3>{title}</h3>
      {props.selectedCampaignName ? (
        <p>Available {title.toLowerCase()} for {props.selectedCampaignName}.</p>
      ) : (
        <p>Select a campaign to load {title.toLowerCase()}.</p>
      )}
      {props.loading ? <p>Loading {title.toLowerCase()}...</p> : null}
      {!props.loading &&
      props.selectedCampaignName &&
      props.contextView === "notes" &&
      props.notes.length === 0 ? (
        <p>No notes associated with this campaign yet.</p>
      ) : null}
      {!props.loading &&
      props.selectedCampaignName &&
      props.contextView === "players" &&
      props.players.length === 0 ? (
        <p>No players associated with this campaign yet.</p>
      ) : null}
      {!props.loading &&
      props.selectedCampaignName &&
      props.contextView === "planes" &&
      props.planes.length === 0 ? (
        <p>No planes associated with this campaign yet.</p>
      ) : null}
      {props.contextView === "notes" && props.notes.length > 0 ? (
        <ul className={styles.noteList}>
          {props.notes.map((note) => (
            <li key={note.id}>
              <button
                type="button"
                className={
                  props.selectedNoteId === note.id
                    ? styles.noteItemActive
                    : styles.noteItem
                }
                onClick={() => props.onSelectNote(note.id)}
              >
                <span className={styles.noteTitle}>{note.title}</span>
                <span className={styles.noteType}>{note.noteType}</span>
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {props.contextView === "players" && props.players.length > 0 ? (
        <ul className={styles.genericList}>
          {props.players.map((player) => (
            <li key={player.id}>
              <button
                type="button"
                className={
                  props.selectedPlayerId === player.id
                    ? styles.genericItemActive
                    : styles.genericItem
                }
                onClick={() => props.onSelectPlayer(player.id)}
              >
                <span className={styles.genericTitle}>{player.name}</span>
                <span className={styles.genericMeta}>PLAYER</span>
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {props.contextView === "planes" && props.planes.length > 0 ? (
        <ul className={styles.genericList}>
          {props.planes.map((plane) => (
            <li key={plane.id}>
              <button
                type="button"
                className={
                  props.selectedPlaneId === plane.id
                    ? styles.genericItemActive
                    : styles.genericItem
                }
                onClick={() => props.onSelectPlane(plane.id)}
              >
                <span className={styles.genericTitle}>{plane.name}</span>
                {plane.description ? (
                  <span className={styles.genericMeta}>{plane.description}</span>
                ) : null}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {props.contextView === "planes" ? (
        <button type="button" className={styles.planeCreateButton} onClick={props.onCreatePlane}>
          + New Plane
        </button>
      ) : null}
    </aside>
  );
}
