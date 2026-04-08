import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { HomePage } from "../pages/HomePage";

describe("HomePage", () => {
    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("renders the oauth landing page title", async () => {
        vi.stubGlobal(
            "fetch",
            vi.fn().mockResolvedValue({
                ok: true,
                json: async () => ({ authenticated: false }),
            }),
        );

        render(
            <MemoryRouter>
                <HomePage />
            </MemoryRouter>,
        );

        await waitFor(() => {
            expect(
                screen.getByText(/Sign in with Warcraft Logs/i),
            ).toBeInTheDocument();
        });
    });
});
