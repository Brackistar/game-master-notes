import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { CampaignViewModel } from "../../models/communication/campaign_models";
import styles from "./CampaignPanel.module.css";

type SortableCampaignRowProps = {
  campaign: CampaignViewModel;
  selected: boolean;
  onSelect: (id: string) => void;
};

export function SortableCampaignRow(props: SortableCampaignRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({
    id: props.campaign.id
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition
  };

  return (
    <li ref={setNodeRef} style={style} className={styles.listItem}>
      <button
        type="button"
        className={props.selected ? styles.itemButtonSelected : styles.itemButton}
        onClick={() => props.onSelect(props.campaign.id)}
      >
        <span className={styles.itemText} title={props.campaign.name}>
          {props.campaign.name}
        </span>
      </button>
      <button
        type="button"
        aria-label={`Drag ${props.campaign.name}`}
        className={styles.dragHandle}
        {...attributes}
        {...listeners}
      >
        ::
      </button>
    </li>
  );
}
