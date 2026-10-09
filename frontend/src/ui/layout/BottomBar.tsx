import styles from "./BottomBar.module.css";

export type CampaignContextView = "notes" | "players" | "planes";

type BottomBarProps = {
  activeContextView: CampaignContextView;
  onChangeContextView: (view: CampaignContextView) => void;
  onQuickAddNote: () => void;
};

export function BottomBar(props: BottomBarProps) {
  const campaignContextActions: { label: string; value: CampaignContextView }[] = [
    { label: "Campaign Notes", value: "notes" },
    { label: "Campaign Players", value: "players" },
    { label: "Campaign Planes", value: "planes" }
  ];

  return (
    <footer className={styles.bottomBar}>
      <section className={styles.leftSection} aria-label="Global search tools">
        <label className={styles.searchLabel}>
          <span>Global Search</span>
          <input type="text" placeholder="Search planes, worlds, campaigns, notes..." />
        </label>
      </section>

      <section className={styles.centerSection} aria-label="Context actions">
        <div className={styles.contextActionGroup}>
          {campaignContextActions.map((action) => (
            <button
              key={action.value}
              type="button"
              aria-pressed={props.activeContextView === action.value}
              className={
                props.activeContextView === action.value
                  ? styles.contextActionActive
                  : styles.contextAction
              }
              onClick={() => props.onChangeContextView(action.value)}
            >
              {action.label}
            </button>
          ))}
        </div>
      </section>

      <section className={styles.rightSection} aria-label="Quick actions">
        <button type="button" className={styles.quickAction} onClick={props.onQuickAddNote}>
          Quick Add Note
        </button>
      </section>
    </footer>
  );
}
