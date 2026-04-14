import styles from "./RightPanel.module.css";

export function RightPanel() {
  return (
    <aside className={styles.rightPanel} aria-label="Context panel">
      <h3>Context Panel</h3>
      <p>Reserved for context actions and details in next iterations.</p>
    </aside>
  );
}
