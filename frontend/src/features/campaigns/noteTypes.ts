export const noteTypeOptions = [
  "general",
  "summary_note",
  "map",
  "character",
  "location"
] as const;

export type NoteTypeOption = (typeof noteTypeOptions)[number];
