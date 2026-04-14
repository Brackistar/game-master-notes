import { useMemo, useState } from "react";
import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent
} from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy
} from "@dnd-kit/sortable";
import type { CampaignViewModel } from "./model";
import { SortableCampaignRow } from "./SortableCampaignRow";
import styles from "./CampaignPanel.module.css";

type CampaignPanelProps = {
  campaigns: CampaignViewModel[];
  selectedCampaignId: string | null;
  onSelectCampaign: (id: string) => void;
  onReorderCampaigns: (idsInOrder: string[]) => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  onAddCampaign: () => void;
  onDeleteCampaign: (id: string) => void;
};

export function CampaignPanel(props: CampaignPanelProps) {
  const [deleteCandidateId, setDeleteCandidateId] = useState<string | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  const deleteCandidate = useMemo(
    () => props.campaigns.find((item) => item.id === deleteCandidateId) ?? null,
    [props.campaigns, deleteCandidateId]
  );

  const filteredCampaigns = useMemo(() => {
    const query = props.searchQuery.trim().toLowerCase();
    if (query === "") {
      return props.campaigns;
    }
    return props.campaigns.filter((campaign) =>
      campaign.name.toLowerCase().includes(query)
    );
  }, [props.campaigns, props.searchQuery]);

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }
    const oldIndex = filteredCampaigns.findIndex((item) => item.id === active.id);
    const newIndex = filteredCampaigns.findIndex((item) => item.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }
    const nextVisible = arrayMove(filteredCampaigns, oldIndex, newIndex).map(
      (item) => item.id
    );
    const visibleSet = new Set(filteredCampaigns.map((item) => item.id));
    const allIds = props.campaigns.map((item) => item.id);
    const positions = allIds
      .map((id, index) => ({ id, index }))
      .filter((item) => visibleSet.has(item.id))
      .map((item) => item.index);
    const merged = [...allIds];
    positions.forEach((position, index) => {
      merged[position] = nextVisible[index];
    });
    props.onReorderCampaigns(merged);
  };

  return (
    <aside className={styles.leftPanel} aria-label="Campaign panel">
      <div className={styles.toolbar}>
        <button
          type="button"
          aria-label="Add campaign"
          title="Add campaign"
          className={styles.iconButton}
          onClick={props.onAddCampaign}
        >
          +
        </button>
        <input
          aria-label="Search campaigns"
          placeholder="Search..."
          value={props.searchQuery}
          onChange={(event) => props.onSearchQueryChange(event.target.value)}
        />
        <button
          type="button"
          aria-label="Delete campaign"
          title="Delete campaign"
          className={styles.iconButton}
          onClick={() => {
            if (props.selectedCampaignId) {
              setDeleteCandidateId(props.selectedCampaignId);
            }
          }}
          disabled={!props.selectedCampaignId}
        >
          🗑
        </button>
      </div>

      <div className={styles.listContainer}>
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={filteredCampaigns.map((item) => item.id)}
            strategy={verticalListSortingStrategy}
          >
            <ul className={styles.campaignList}>
              {filteredCampaigns.map((campaign) => (
                <SortableCampaignRow
                  key={campaign.id}
                  campaign={campaign}
                  selected={campaign.id === props.selectedCampaignId}
                  onSelect={props.onSelectCampaign}
                />
              ))}
            </ul>
          </SortableContext>
        </DndContext>
      </div>

      {deleteCandidate ? (
        <div className={styles.deleteModalBackdrop} role="dialog" aria-modal="true">
          <div className={styles.deleteModal}>
            <p>
              Delete campaign <strong>{deleteCandidate.name}</strong>?
            </p>
            <div className={styles.deleteActions}>
              <button type="button" onClick={() => setDeleteCandidateId(null)}>
                Cancel
              </button>
              <button
                type="button"
                onClick={() => {
                  props.onDeleteCampaign(deleteCandidate.id);
                  setDeleteCandidateId(null);
                }}
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </aside>
  );
}
