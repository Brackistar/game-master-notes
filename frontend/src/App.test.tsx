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

    expect(await screen.findByText("Crimson Tides")).toBeInTheDocument();
    expect(screen.getByText("Campaign Notes Workspace")).toBeInTheDocument();
  });
});
