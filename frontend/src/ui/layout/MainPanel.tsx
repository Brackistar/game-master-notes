import styles from "./MainPanel.module.css";

type MainPanelProps = {
  mode: "empty" | "create-campaign";
};

export function MainPanel(props: MainPanelProps) {
  if (props.mode === "create-campaign") {
    return (
      <main className={styles.mainPanel}>
        <div className={styles.centerCard}>
          <h2>Create Campaign</h2>
          <p>This first baseline switches the center workspace into create mode.</p>
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
