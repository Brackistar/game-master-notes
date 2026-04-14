import styles from "./BottomBar.module.css";

export function BottomBar() {
  return (
    <footer className={styles.bottomBar}>
      <label className={styles.searchLabel}>
        <span>Global Search</span>
        <input type="text" placeholder="Search worlds, campaigns, notes..." />
      </label>
      <button type="button">Quick Add Note</button>
    </footer>
  );
}
