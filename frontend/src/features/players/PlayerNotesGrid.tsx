import { useEffect, useMemo, useRef, useState } from "react";
import type { PlayerNoteCard } from "../../models/communication/campaign_models";
import styles from "./PlayerNotesGrid.module.css";

const minColumns = 3;
const maxColumns = 5;
const cardMinWidth = 220;
const rowGap = 12;
const colGap = 12;
const cardHeightRatio = 1.20;
const bufferRows = 2;

type PlayerNotesGridProps = {
  notes: PlayerNoteCard[];
  onSelectNote: (noteId: string) => void;
};

export function PlayerNotesGrid(props: PlayerNotesGridProps) {
  const viewportRef = useRef<HTMLDivElement | null>(null);
  const [viewportWidth, setViewportWidth] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(0);
  const [scrollTop, setScrollTop] = useState(0);

  useEffect(() => {
    const element = viewportRef.current;
    if (!element) {
      return;
    }

    if (typeof ResizeObserver === "undefined") {
      setViewportWidth(element.clientWidth || 960);
      setViewportHeight(element.clientHeight || 560);
      return;
    }

    const observer = new ResizeObserver(() => {
      setViewportWidth(element.clientWidth);
      setViewportHeight(element.clientHeight);
    });
    observer.observe(element);

    setViewportWidth(element.clientWidth);
    setViewportHeight(element.clientHeight);

    return () => observer.disconnect();
  }, []);

  const columns = useMemo(() => {
    if (viewportWidth <= 0) {
      return minColumns;
    }
    const estimate = Math.floor((viewportWidth + colGap) / (cardMinWidth + colGap));
    return Math.max(minColumns, Math.min(maxColumns, estimate));
  }, [viewportWidth]);

  const cardWidth =
    viewportWidth > 0
      ? (viewportWidth - colGap * (columns - 1)) / columns
      : cardMinWidth;
  const cardHeight = Math.round(cardWidth * cardHeightRatio);
  const rowHeight = cardHeight + rowGap;
  const totalRows = Math.ceil(props.notes.length / columns);
  const totalHeight = totalRows * rowHeight;

  const firstVisibleRow = Math.max(0, Math.floor(scrollTop / rowHeight));
  const visibleRowCount =
    viewportHeight > 0 ? Math.ceil(viewportHeight / rowHeight) : minColumns;
  const startRow = Math.max(0, firstVisibleRow - bufferRows);
  const endRow = Math.min(totalRows, firstVisibleRow + visibleRowCount + bufferRows);

  const visibleNotes = useMemo(() => {
    const items: { note: PlayerNoteCard; top: number; left: number; width: number }[] = [];
    for (let row = startRow; row < endRow; row += 1) {
      for (let col = 0; col < columns; col += 1) {
        const index = row * columns + col;
        if (index >= props.notes.length) {
          break;
        }
        items.push({
          note: props.notes[index],
          top: row * rowHeight,
          left: col * (cardWidth + colGap),
          width: cardWidth
        });
      }
    }
    return items;
  }, [cardWidth, columns, endRow, props.notes, rowHeight, startRow]);

  return (
    <div
      ref={viewportRef}
      className={styles.viewport}
      onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
    >
      <div className={styles.canvas} style={{ height: `${totalHeight}px` }}>
        {visibleNotes.map((item) => (
          <button
            key={item.note.id}
            type="button"
            className={styles.noteCard}
            style={{
              top: `${item.top}px`,
              left: `${item.left}px`,
              width: `${item.width}px`,
              height: `${cardHeight}px`
            }}
            onClick={() => props.onSelectNote(item.note.id)}
          >
            <span className={styles.cardTitle}>{item.note.title}</span>
            <span className={styles.cardType}>{item.note.noteType.toUpperCase()}</span>
            <span className={styles.cardContent}>{toPreview(item.note.contentMd)}</span>
          </button>
        ))}
      </div>
    </div>
  );
}

function toPreview(markdown: string): string {
  const text = markdown
    .replace(/[#>*`_\-\[\]\(\)]/g, "")
    .replace(/\n+/g, " ")
    .trim();
  return text.length > 180 ? `${text.slice(0, 180)}...` : text;
}
