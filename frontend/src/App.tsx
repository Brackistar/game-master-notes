import { useMemo, useState } from "react";
import { CampaignPanel } from "./features/campaigns/CampaignPanel";
import { BottomBar } from "./ui/layout/BottomBar";
import { MainPanel } from "./ui/layout/MainPanel";
import { RightPanel } from "./ui/layout/RightPanel";
import { initialCampaigns } from "./features/campaigns/mockCampaigns";
import styles from "./ui/layout/AppShell.module.css";

type CenterMode = "empty" | "create-campaign";

export type CampaignItem = {
  id: string;
  name: string;
};

export function App() {
  const [campaigns, setCampaigns] = useState<CampaignItem[]>(initialCampaigns);
  const [selectedCampaignId, setSelectedCampaignId] = useState<string | null>(
    initialCampaigns[0]?.id ?? null
  );
  const [searchQuery, setSearchQuery] = useState("");
  const [centerMode, setCenterMode] = useState<CenterMode>("empty");

  const filteredCampaigns = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (query === "") {
      return campaigns;
    }
    return campaigns.filter((campaign) =>
      campaign.name.toLowerCase().includes(query)
    );
  }, [campaigns, searchQuery]);

  return (
    <div className={styles.appShell}>
      <div className={styles.workspaceRow}>
        <CampaignPanel
          campaigns={filteredCampaigns}
          selectedCampaignId={selectedCampaignId}
          onSelectCampaign={setSelectedCampaignId}
          onReorderCampaigns={setCampaigns}
          searchQuery={searchQuery}
          onSearchQueryChange={setSearchQuery}
          onAddCampaign={() => setCenterMode("create-campaign")}
          onDeleteCampaign={(id) => {
            setCampaigns((previous) => previous.filter((item) => item.id !== id));
            setSelectedCampaignId((current) => (current === id ? null : current));
          }}
        />
        <MainPanel mode={centerMode} />
        <RightPanel />
      </div>
      <BottomBar />
    </div>
  );
}
