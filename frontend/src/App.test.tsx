import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "./App";

describe("App shell", () => {
  it("renders campaign panel and bottom bar controls", () => {
    render(<App />);

    expect(screen.getByLabelText("Campaign panel")).toBeInTheDocument();
    expect(screen.getByLabelText("Context panel")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Search...")).toBeInTheDocument();
    expect(screen.getByText("Quick Add Note")).toBeInTheDocument();
  });

  it("switches center view to create mode when add is clicked", async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: "Add campaign" }));

    expect(screen.getByText("Create Campaign")).toBeInTheDocument();
  });

  it("filters campaign list with case-insensitive contains", async () => {
    const user = userEvent.setup();
    render(<App />);
    await screen.findByText("Ashes of the Ivory Coast");

    await user.type(screen.getByLabelText("Search campaigns"), "glass");

    expect(screen.getByText("The Glass Crown Conspiracy")).toBeInTheDocument();
    expect(screen.queryByText("Ashes of the Ivory Coast")).not.toBeInTheDocument();
  });

  it("creates a campaign from center form", async () => {
    const user = userEvent.setup();
    render(<App />);
    await screen.findByText("The Glass Crown Conspiracy");

    await user.click(screen.getByRole("button", { name: "Add campaign" }));
    await user.type(screen.getByLabelText("Campaign Name"), "Crimson Tides");
    await user.click(screen.getByRole("button", { name: "Create" }));

    const crimsonMatches = await screen.findAllByText("Crimson Tides");
    expect(crimsonMatches.length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "Crimson Tides" })).toBeInTheDocument();
  });

  it("shows note content in center panel when selecting a campaign note", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Ashes of the Ivory Coast" }));
    await user.click(screen.getByRole("button", { name: /Session Zero - Factions/i }));

    expect(await screen.findByText("Factions")).toBeInTheDocument();
    expect(screen.getByText("Salt Cartel")).toBeInTheDocument();
  });

  it("edits a campaign note and shows updated markdown content", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Ashes of the Ivory Coast" }));
    await user.click(screen.getByRole("button", { name: /Session Zero - Factions/i }));
    await user.click(screen.getByRole("button", { name: "Edit" }));

    const titleInput = screen.getByLabelText("Title");
    const contentInput = screen.getByLabelText("Markdown Content");

    await user.clear(titleInput);
    await user.type(titleInput, "Session Zero - Alliances");
    await user.clear(contentInput);
    await user.type(contentInput, "# Alliances{enter}- Tide Blades");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Alliances")).toBeInTheDocument();
    expect(screen.getByText("Tide Blades")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Session Zero - Alliances/i })).toBeInTheDocument();
  });

  it("switches right panel context between notes, players, and planes", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Campaign Players" }));
    expect(await screen.findByText("Elara Stormborne")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Campaign Planes" }));
    const planeMatches = await screen.findAllByText("Prime Material Reach");
    expect(planeMatches.length).toBeGreaterThan(0);

    await user.click(screen.getByRole("button", { name: "Campaign Notes" }));
    expect(await screen.findByRole("button", { name: /Session Zero - Factions/i })).toBeInTheDocument();
  });

  it("shows plane details and associated worlds in center when selecting a campaign plane", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Campaign Planes" }));
    await user.click(screen.getByRole("button", { name: /Prime Material Reach/i }));

    expect(await screen.findByText("Worlds In This Plane")).toBeInTheDocument();
    expect(screen.getByText("Ivory Storm Coast")).toBeInTheDocument();
    expect(screen.getByText("Emberglass Dominion")).toBeInTheDocument();
  });

  it("opens world main view when selecting a world from plane workspace", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Campaign Planes" }));
    await user.click(screen.getByRole("button", { name: /Prime Material Reach/i }));
    await user.click(screen.getByRole("button", { name: /Ivory Storm Coast/i }));

    expect(
      await screen.findByText("World ID: 01HQ8J5QK8W8S6W2A3E0P8W001")
    ).toBeInTheDocument();
    expect(screen.getByText("World Workspace")).toBeInTheDocument();
  });

  it("keeps center panel state unchanged when right panel context changes", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: /Session Zero - Factions/i }));
    expect(await screen.findByText("Factions")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Campaign Players" }));
    expect(screen.getByText("Factions")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Campaign Planes" }));
    expect(screen.getByText("Factions")).toBeInTheDocument();
  });

  it("shows player main view in center when a campaign player is clicked", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Campaign Players" }));
    await user.click(screen.getByRole("button", { name: /Elara Stormborne/i }));

    expect(await screen.findByText("Player ID: PLAYER-001")).toBeInTheDocument();
    expect(screen.getByText("Player Notes")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^Elara Chronicle 1\b/i })).toBeInTheDocument();
  });

  it("virtualizes player notes grid instead of rendering all notes", async () => {
    const user = userEvent.setup();
    render(<App />);

    await screen.findByText("Ashes of the Ivory Coast");
    await user.click(screen.getByRole("button", { name: "Campaign Players" }));
    await user.click(screen.getByRole("button", { name: /Elara Stormborne/i }));

    expect(await screen.findByRole("button", { name: /^Elara Chronicle 1\b/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Elara Chronicle 42\b/i })).not.toBeInTheDocument();
  });
});
