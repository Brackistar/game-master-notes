import { useMemo, useState, type FormEvent } from "react";
import styles from "../../ui/layout/MainPanel.module.css";

type CreateCampaignFormProps = {
  onCreate: (name: string) => Promise<void>;
  onCancel: () => void;
};

export function CreateCampaignForm(props: CreateCampaignFormProps) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const validationError = useMemo(() => validateCampaignName(name), [name]);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (validationError) {
      setError(validationError);
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      await props.onCreate(name.trim());
      setName("");
    } catch {
      setError("Unable to create campaign right now.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <label htmlFor="campaign-name" className={styles.formLabel}>
        Campaign Name
      </label>
      <input
        id="campaign-name"
        value={name}
        onChange={(event) => setName(event.target.value)}
        placeholder="Type a campaign name..."
        maxLength={50}
        autoFocus
      />
      <p className={styles.formHint}>3 to 50 characters, trimmed before saving.</p>
      {error ? (
        <p role="alert" className={styles.formError}>
          {error}
        </p>
      ) : null}
      <div className={styles.formActions}>
        <button type="button" onClick={props.onCancel} disabled={submitting}>
          Cancel
        </button>
        <button type="submit" disabled={submitting}>
          {submitting ? "Creating..." : "Create"}
        </button>
      </div>
    </form>
  );
}

function validateCampaignName(value: string): string | null {
  const normalized = value.trim();
  if (normalized.length === 0) {
    return "Campaign name is required.";
  }
  if (normalized.length < 3) {
    return "Campaign name must have at least 3 characters.";
  }
  if (normalized.length > 50) {
    return "Campaign name must have at most 50 characters.";
  }
  return null;
}
