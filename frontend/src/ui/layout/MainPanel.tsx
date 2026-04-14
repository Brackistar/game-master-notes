import { CreateCampaignForm } from "../../features/campaigns/CreateCampaignForm";
import styles from "./MainPanel.module.css";

type MainPanelProps = {
  mode: "empty" | "create-campaign";
  onCreateCampaign: (name: string) => Promise<void>;
  onCancelCreateCampaign: () => void;
};

export function MainPanel(props: MainPanelProps) {
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

  return (
    <main className={styles.mainPanel}>
      <div className={styles.centerCard}>
        <h2>Campaign Notes Workspace</h2>
        <p>Select a campaign on the left or click Add to start a new campaign.</p>
      </div>
    </main>
  );
}
