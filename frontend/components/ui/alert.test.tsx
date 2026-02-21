import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Alert, AlertTitle, AlertDescription } from "./alert";

describe("Alert", () => {
  it("renders correctly with default variant", () => {
    render(
      <Alert>
        <AlertTitle>Heads up!</AlertTitle>
        <AlertDescription>This is a default alert.</AlertDescription>
      </Alert>,
    );

    const alert = screen.getByRole("alert");
    expect(alert).toBeInTheDocument();
    expect(alert).toHaveClass("bg-card");
    expect(screen.getByText("Heads up!")).toBeInTheDocument();
    expect(screen.getByText("This is a default alert.")).toBeInTheDocument();
  });

  it("renders destructive variant", () => {
    render(
      <Alert variant="destructive">
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>Something went wrong.</AlertDescription>
      </Alert>,
    );

    const alert = screen.getByRole("alert");
    expect(alert).toHaveClass("text-destructive");
    expect(alert).toHaveClass("bg-card");
  });

  it("applies custom className", () => {
    render(
      <Alert className="custom-alert">
        <AlertTitle>Title</AlertTitle>
        <AlertDescription>Desc</AlertDescription>
      </Alert>,
    );

    const alert = screen.getByRole("alert");
    expect(alert).toHaveClass("custom-alert");
  });
});
