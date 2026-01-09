import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { BranchSelect } from "../BranchSelect";

// Mock the API module
vi.mock("@/lib/api", () => ({
  repositoryApi: {
    listBranches: vi.fn(),
  },
}));

import { repositoryApi } from "@/lib/api";
const mockRepositoryApi = vi.mocked(repositoryApi);

describe("BranchSelect Component", () => {
  const mockBranches = ["main", "develop", "feature/new-feature"];
  const mockAccessToken = "test-access-token";

  // New self-contained repository model (no git_provider_id)
  const mockRepository = {
    id: 1,
    organization_id: 1,
    provider_type: "github",
    provider_base_url: "https://github.com",
    clone_url: "https://github.com/org/my-repo.git",
    external_id: "12345",
    name: "my-repo",
    full_path: "org/my-repo",
    default_branch: "main",
    visibility: "organization",
    is_active: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockRepositoryApi.listBranches.mockResolvedValue({ branches: mockBranches });
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe("no repository selected", () => {
    it("should show disabled select with message", () => {
      render(
        <BranchSelect
          repositoryId={null}
          value=""
          onChange={() => {}}
        />
      );

      expect(screen.getByText("Select a repository first")).toBeInTheDocument();
      expect(screen.getByRole("combobox")).toBeDisabled();
    });
  });

  describe("no access token provided", () => {
    it("should switch to manual mode when no access token", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          allowManualInput
        />
      );

      // Without access token, it should go to manual mode
      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });

      // API should not be called without access token
      expect(mockRepositoryApi.listBranches).not.toHaveBeenCalled();
    });
  });

  describe("loading state", () => {
    it("should show loading text while fetching branches", () => {
      mockRepositoryApi.listBranches.mockImplementation(() => new Promise(() => {}));

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      expect(screen.getByText("Loading branches...")).toBeInTheDocument();
    });
  });

  describe("rendering branches", () => {
    it("should render branches after loading", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("main")).toBeInTheDocument();
      });

      expect(screen.getByText("develop")).toBeInTheDocument();
      expect(screen.getByText("feature/new-feature")).toBeInTheDocument();
    });

    it("should mark default branch", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          repository={mockRepository}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("main (default)")).toBeInTheDocument();
      });
    });

    it("should show placeholder when provided", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          placeholder="Pick a branch..."
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("Pick a branch...")).toBeInTheDocument();
      });
    });
  });

  describe("auto-selection", () => {
    it("should auto-select default branch when available", async () => {
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          repository={mockRepository}
          value=""
          onChange={handleChange}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(handleChange).toHaveBeenCalledWith("main");
      });
    });

    it("should auto-select first branch when default not available", async () => {
      const handleChange = vi.fn();
      mockRepositoryApi.listBranches.mockResolvedValue({
        branches: ["develop", "feature/x"],
      });

      render(
        <BranchSelect
          repositoryId={1}
          repository={mockRepository}
          value=""
          onChange={handleChange}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(handleChange).toHaveBeenCalledWith("develop");
      });
    });

    it("should not auto-select when value is already set", async () => {
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          repository={mockRepository}
          value="develop"
          onChange={handleChange}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        // Look for option value instead of text due to "(default)" suffix
        expect(screen.getByRole("combobox")).toBeInTheDocument();
        expect(screen.getByRole("option", { name: /main/ })).toBeInTheDocument();
      });

      // Should not have changed the value
      expect(handleChange).not.toHaveBeenCalled();
    });
  });

  describe("selection", () => {
    it("should call onChange when branch selected", async () => {
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          value="main"
          onChange={handleChange}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("develop")).toBeInTheDocument();
      });

      fireEvent.change(screen.getByRole("combobox"), { target: { value: "develop" } });

      expect(handleChange).toHaveBeenCalledWith("develop");
    });

    it("should show selected branch", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value="develop"
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole("combobox")).toHaveValue("develop");
      });
    });
  });

  describe("manual input mode", () => {
    it("should switch to manual mode when no branches available", async () => {
      mockRepositoryApi.listBranches.mockResolvedValue({ branches: [] });

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });
    });

    it("should switch to manual mode on API error", async () => {
      mockRepositoryApi.listBranches.mockRejectedValue(new Error("API error"));

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });
    });

    it("should allow typing branch name in manual mode", async () => {
      mockRepositoryApi.listBranches.mockResolvedValue({ branches: [] });
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={handleChange}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });

      fireEvent.change(screen.getByPlaceholderText("Enter branch name (e.g., main)"), {
        target: { value: "my-branch" },
      });

      expect(handleChange).toHaveBeenCalledWith("my-branch");
    });

    it("should switch to manual mode when Enter manually option selected", async () => {
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          value="main"
          onChange={handleChange}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("Enter manually...")).toBeInTheDocument();
      });

      fireEvent.change(screen.getByRole("combobox"), { target: { value: "__manual__" } });

      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });
    });

    it("should show List button to go back to select mode", async () => {
      mockRepositoryApi.listBranches.mockResolvedValue({ branches: mockBranches });
      const handleChange = vi.fn();

      render(
        <BranchSelect
          repositoryId={1}
          value="main"
          onChange={handleChange}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("Enter manually...")).toBeInTheDocument();
      });

      // Switch to manual mode
      fireEvent.change(screen.getByRole("combobox"), { target: { value: "__manual__" } });

      await waitFor(() => {
        expect(screen.getByText("List")).toBeInTheDocument();
      });

      // Go back to list
      fireEvent.click(screen.getByText("List"));

      await waitFor(() => {
        expect(screen.getByRole("combobox")).toBeInTheDocument();
        expect(screen.getByText("main")).toBeInTheDocument();
      });
    });
  });

  describe("disabled state", () => {
    it("should disable select when disabled prop is true", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          disabled
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole("combobox")).toBeDisabled();
      });
    });

    it("should disable input in manual mode when disabled", async () => {
      mockRepositoryApi.listBranches.mockResolvedValue({ branches: [] });

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          disabled
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeDisabled();
      });
    });
  });

  describe("error handling without allowManualInput", () => {
    it("should show error message when API fails", async () => {
      mockRepositoryApi.listBranches.mockRejectedValue(new Error("Network error"));

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          allowManualInput={false}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("Failed to load branches")).toBeInTheDocument();
      });
    });

    it("should switch to manual mode on error when allowManualInput", async () => {
      mockRepositoryApi.listBranches.mockRejectedValue(new Error("Network error"));

      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          allowManualInput
          accessToken={mockAccessToken}
        />
      );

      // When allowManualInput is true, it should go to manual mode
      await waitFor(() => {
        expect(screen.getByPlaceholderText("Enter branch name (e.g., main)")).toBeInTheDocument();
      });
    });
  });

  describe("repository change", () => {
    it("should reload branches when repositoryId changes", async () => {
      const { rerender } = render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(mockRepositoryApi.listBranches).toHaveBeenCalledWith(1, mockAccessToken);
      });

      mockRepositoryApi.listBranches.mockResolvedValue({
        branches: ["master", "staging"],
      });

      rerender(
        <BranchSelect
          repositoryId={2}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(mockRepositoryApi.listBranches).toHaveBeenCalledWith(2, mockAccessToken);
      });
    });

    it("should clear branches when repository is deselected", async () => {
      const { rerender } = render(
        <BranchSelect
          repositoryId={1}
          value="main"
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByText("main")).toBeInTheDocument();
      });

      rerender(
        <BranchSelect
          repositoryId={null}
          value=""
          onChange={() => {}}
          accessToken={mockAccessToken}
        />
      );

      expect(screen.getByText("Select a repository first")).toBeInTheDocument();
    });
  });

  describe("className prop", () => {
    it("should apply custom className to select", async () => {
      render(
        <BranchSelect
          repositoryId={1}
          value=""
          onChange={() => {}}
          className="custom-class"
          accessToken={mockAccessToken}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole("combobox")).toHaveClass("custom-class");
      });
    });
  });
});
