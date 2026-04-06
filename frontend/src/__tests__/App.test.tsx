import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { HomePage } from "../pages/HomePage";

describe("HomePage", () => {
    it("renders the landing page title", () => {
        render(
            <MemoryRouter>
                <HomePage />
            </MemoryRouter>,
        );

        expect(
            screen.getByText(/Paste your Warcraft Logs report URL/i),
        ).toBeInTheDocument();
    });
});
