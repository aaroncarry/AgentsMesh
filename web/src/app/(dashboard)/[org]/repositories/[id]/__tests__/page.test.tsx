import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import RepositoryDetailPage from "../page";

// Mock next/navigation
const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useParams: () => ({ id: "1" }),
  useRouter: () => ({ push: mockPush }),
}));

// Mock next/link
vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

// Mock the API modules
vi.mock("@/lib/api", () => ({
  repositoryApi: {
    get: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    listBranches: vi.fn(),
    syncBranches: vi.fn(),
    setupWebhook: vi.fn(),
  },
  gitProviderApi: {
    get: vi.fn(),
  },
}));

import { repositoryApi, gitProviderApi } from "@/lib/api";
const mockRepositoryApi = vi.mocked(repositoryApi);
const mockGitProviderApi = vi.mocked(gitProviderApi);

describe("RepositoryDetailPage", () => {
  const mockRepository = {
    id: 1,
    organization_id: 1,
    git_provider_id: 1,
    external_id: "12345",
    name: "my-repo",
    full_path: "org/my-repo",
    default_branch: "main",
    ticket_prefix: "PROJ",
    is_active: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  };

  const mockGitProvider = {
    id: 1,
    organization_id: 1,
    name: "GitHub",
    provider_type: "github",
    base_url: "https://github.com",
    is_active: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  };

  const mockBranches = ["main", "develop", "feature/new-feature"];

  beforeEach(() => {
    vi.clearAllMocks();
    mockRepositoryApi.get.mockResolvedValue({ repository: mockRepository });
    mockGitProviderApi.get.mockResolvedValue({ git_provider: mockGitProvider });
    mockRepositoryApi.listBranches.mockResolvedValue({ branches: mockBranches });
    mockRepositoryApi.syncBranches.mockResolvedValue({
      branches: mockBranches,
      message: "Synced",
    });
    mockRepositoryApi.setupWebhook.mockResolvedValue({
      message: "Webhook created",
      webhook_url: "https://api.example.com/webhooks/123",
    });
    mockRepositoryApi.delete.mockResolvedValue({ message: "Deleted" });
    mockRepositoryApi.update.mockResolvedValue({ repository: mockRepository });
    // Mock window.confirm and alert
    vi.spyOn(window, "confirm").mockReturnValue(true);
    vi.spyOn(window, "alert").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("loading state", () => {
    it("should show loading spinner initially", () => {
      mockRepositoryApi.get.mockImplementation(() => new Promise(() => {}));

      render(<RepositoryDetailPage />);

      expect(document.querySelector(".animate-spin")).toBeTruthy();
    });
  });

  describe("not found state", () => {
    it("should show not found message when repository not found", async () => {
      mockRepositoryApi.get.mockRejectedValue(new Error("Not found"));

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Repository not found")).toBeInTheDocument();
      });
    });

    it("should show back button when not found", async () => {
      mockRepositoryApi.get.mockRejectedValue(new Error("Not found"));

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Back to Repositories")).toBeInTheDocument();
      });
    });
  });

  describe("rendering", () => {
    it("should render repository name", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        // Multiple instances of name appear (header, breadcrumb, details)
        expect(screen.getAllByText("my-repo").length).toBeGreaterThanOrEqual(1);
      });
    });

    it("should render repository full path", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        // Multiple instances of path appear (header, details)
        expect(screen.getAllByText("org/my-repo").length).toBeGreaterThanOrEqual(1);
      });
    });

    it("should render Edit and Delete buttons", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Edit")).toBeInTheDocument();
        expect(screen.getByText("Delete")).toBeInTheDocument();
      });
    });

    it("should render breadcrumb", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByRole("link", { name: "Repositories" })).toBeInTheDocument();
      });
    });

    it("should render tabs", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Information")).toBeInTheDocument();
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });
    });
  });

  describe("information tab", () => {
    it("should show repository details section", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Repository Details")).toBeInTheDocument();
      });
    });

    it("should show default branch", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Default Branch")).toBeInTheDocument();
        expect(screen.getByText("main")).toBeInTheDocument();
      });
    });

    it("should show ticket prefix", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Ticket Prefix")).toBeInTheDocument();
        expect(screen.getByText("PROJ")).toBeInTheDocument();
      });
    });

    it("should show active status", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Status")).toBeInTheDocument();
        expect(screen.getByText("Active")).toBeInTheDocument();
      });
    });

    it("should show git provider info", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Git Provider")).toBeInTheDocument();
        expect(screen.getByText("GitHub")).toBeInTheDocument();
      });
    });

    it("should show actions section", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Actions")).toBeInTheDocument();
        expect(screen.getByText("Setup Webhook")).toBeInTheDocument();
        expect(screen.getByText("Sync Branches")).toBeInTheDocument();
      });
    });
  });

  describe("branches tab", () => {
    it("should switch to branches tab", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Branches"));

      await waitFor(() => {
        expect(mockRepositoryApi.listBranches).toHaveBeenCalledWith(1);
      });
    });

    it("should render branches list", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Branches"));

      await waitFor(() => {
        expect(screen.getByText("main")).toBeInTheDocument();
        expect(screen.getByText("develop")).toBeInTheDocument();
        expect(screen.getByText("feature/new-feature")).toBeInTheDocument();
      });
    });

    it("should show default badge on default branch", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Branches"));

      await waitFor(() => {
        expect(screen.getByText("default")).toBeInTheDocument();
      });
    });

    it("should have sync button in branches tab", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Branches"));

      await waitFor(() => {
        expect(screen.getByText("Sync")).toBeInTheDocument();
      });
    });
  });

  describe("delete functionality", () => {
    it("should call delete API when Delete clicked and confirmed", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Delete")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Delete"));

      expect(window.confirm).toHaveBeenCalled();
      await waitFor(() => {
        expect(mockRepositoryApi.delete).toHaveBeenCalledWith(1);
      });
    });

    it("should navigate to repositories list after delete", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Delete")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Delete"));

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith("../repositories");
      });
    });

    it("should not delete when cancelled", async () => {
      vi.spyOn(window, "confirm").mockReturnValue(false);

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Delete")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Delete"));

      expect(mockRepositoryApi.delete).not.toHaveBeenCalled();
    });
  });

  describe("webhook setup", () => {
    it("should call setupWebhook API when button clicked", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Setup Webhook")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Setup Webhook"));

      await waitFor(() => {
        expect(mockRepositoryApi.setupWebhook).toHaveBeenCalledWith(1);
      });
    });

    it("should show alert with webhook info", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Setup Webhook")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Setup Webhook"));

      await waitFor(() => {
        expect(window.alert).toHaveBeenCalledWith(
          expect.stringContaining("Webhook created")
        );
      });
    });
  });

  describe("sync branches", () => {
    it("should call syncBranches API when button clicked", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Sync Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Sync Branches"));

      await waitFor(() => {
        expect(mockRepositoryApi.syncBranches).toHaveBeenCalledWith(1);
      });
    });
  });

  describe("edit modal", () => {
    it("should open edit modal when Edit clicked", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Edit")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Edit"));

      expect(screen.getByText("Edit Repository")).toBeInTheDocument();
    });

    it("should close edit modal when Cancel clicked", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Edit")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Edit"));
      fireEvent.click(screen.getByText("Cancel"));

      await waitFor(() => {
        expect(screen.queryByText("Edit Repository")).not.toBeInTheDocument();
      });
    });

    it("should call update API when save clicked", async () => {
      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Edit")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Edit"));

      // Change the name
      const nameInput = screen.getByDisplayValue("my-repo");
      fireEvent.change(nameInput, { target: { value: "updated-repo" } });

      fireEvent.click(screen.getByText("Save Changes"));

      await waitFor(() => {
        expect(mockRepositoryApi.update).toHaveBeenCalledWith(1, expect.objectContaining({
          name: "updated-repo",
        }));
      });
    });
  });

  describe("inactive repository", () => {
    it("should show Inactive badge for inactive repository", async () => {
      mockRepositoryApi.get.mockResolvedValue({
        repository: { ...mockRepository, is_active: false },
      });

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        // Multiple "Inactive" elements: header badge and status section
        expect(screen.getAllByText("Inactive").length).toBeGreaterThanOrEqual(1);
      });
    });
  });

  describe("SSH provider", () => {
    it("should show SSH provider type in provider info", async () => {
      mockGitProviderApi.get.mockResolvedValue({
        git_provider: { ...mockGitProvider, provider_type: "ssh", name: "SSH Provider" },
      });

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        // Should show provider type as "ssh" (CSS handles capitalization)
        expect(screen.getByText("ssh")).toBeInTheDocument();
      });
    });

    it("should show message about unavailable branches for SSH", async () => {
      mockGitProviderApi.get.mockResolvedValue({
        git_provider: { ...mockGitProvider, provider_type: "ssh" },
      });
      mockRepositoryApi.listBranches.mockResolvedValue({ branches: [] });

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Branches")).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText("Branches"));

      await waitFor(() => {
        expect(screen.getByText("Branch listing not available for SSH providers")).toBeInTheDocument();
      });
    });
  });

  describe("error handling", () => {
    it("should handle git provider fetch error gracefully", async () => {
      mockGitProviderApi.get.mockRejectedValue(new Error("Not found"));

      render(<RepositoryDetailPage />);

      await waitFor(() => {
        expect(screen.getByText("Provider information not available")).toBeInTheDocument();
      });
    });
  });
});
