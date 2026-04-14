import { useCallback, useEffect, useMemo, useState } from "react";
import { CampaignPanel } from "./features/campaigns/CampaignPanel";
import { BottomBar } from "./ui/layout/BottomBar";
import { MainPanel } from "./ui/layout/MainPanel";
import { RightPanel } from "./ui/layout/RightPanel";
import { createCampaignDataSource } from "./features/campaigns/data/createCampaignDataSource";
import type { CampaignViewModel } from "./features/campaigns/model";
import styles from "./ui/layout/AppShell.module.css";

type CenterMode = "empty" | "create-campaign";

export function App() {
  const campaignDataSource = useMemo(() => createCampaignDataSource(), []);
  const [campaigns, setCampaigns] = useState<CampaignViewModel[]>([]);
  const [selectedCampaignId, setSelectedCampaignId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [centerMode, setCenterMode] = useState<CenterMode>("empty");
  const [loadError, setLoadError] = useState<string | null>(null);

  const selectedCampaignExists = useMemo(
    () =>
      selectedCampaignId !== null &&
      campaigns.some((campaign) => campaign.id === selectedCampaignId),
    [campaigns, selectedCampaignId]
  );

  useEffect(() => {
    void (async () => {
      try {
        setLoadError(null);
        const items = await campaignDataSource.listCampaigns();
        setCampaigns(items);
        setSelectedCampaignId(items[0]?.id ?? null);
      } catch {
        setLoadError("Unable to load campaigns.");
      }
    })();
  }, [campaignDataSource]);

  const handleReorderCampaigns = useCallback(
    async (idsInOrder: string[]) => {
      const byId = new Map(campaigns.map((item) => [item.id, item]));
      const sorted = idsInOrder
        .map((id) => byId.get(id))
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

  const handleCreateCampaign = useCallback(async (name: string) => {
    const created = await campaignDataSource.createCampaign({ name });
    setCampaigns((previous) => [created, ...previous]);
    setSelectedCampaignId(created.id);
    setCenterMode("empty");
    setSearchQuery("");
  }, [campaignDataSource]);

  const handleDeleteCampaign = useCallback(async (id: string) => {
    try {
      await campaignDataSource.deleteCampaign(id);
      setCampaigns((previous) => previous.filter((item) => item.id !== id));
      setSelectedCampaignId((current) => (current === id ? null : current));
    } catch {
      setLoadError("Unable to delete campaign.");
    }
  }, [campaignDataSource]);

  return (
    <div className={styles.appShell}>
      <div className={styles.workspaceRow}>
        <CampaignPanel
          campaigns={campaigns}
          selectedCampaignId={selectedCampaignId}
          onSelectCampaign={setSelectedCampaignId}
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
        />
        <RightPanel />
      </div>
      {loadError ? <div className={styles.errorBar}>{loadError}</div> : null}
      {!selectedCampaignExists && campaigns.length > 0 ? (
        <div className={styles.errorBar}>Select a campaign to continue.</div>
      ) : null}
      <BottomBar />
    </div>
  );
}
