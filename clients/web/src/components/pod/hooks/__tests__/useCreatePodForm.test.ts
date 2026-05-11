import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { getPodService, getUserCredentialService } from "@/lib/wasm-core";
import { POD_MODE_ACP } from "@/lib/pod-modes";

vi.mock("@/stores/podCreation", () => ({
  usePodCreationStore: () => ({
    lastAgentSlug: null,
    lastRepositoryId: null,
    lastCredentialProfileId: null,
    lastBranchName: null,
    setLastChoices: vi.fn(),
    clearLastChoices: vi.fn(),
    _hasHydrated: true,
    setHasHydrated: vi.fn(),
  }),
}));

import { useCreatePodForm, RUNNER_HOST_PROFILE_ID } from "../useCreatePodForm";

const mockAgents = [
  { name: "Claude Code", slug: "claude-code", is_builtin: true, is_active: true },
];
const factoryAgent = {
  name: "Factory CLI",
  slug: "factory-cli",
  is_builtin: true,
  is_active: true,
  supported_modes: "pty,acp",
};

const mockCreatePod = vi.fn();
const mockListCredentials = vi.fn();

function setupMocks() {
  // podState is a stable singleton — mutate its create_pod
  const podSvc = getPodService();
  (podSvc as unknown as Record<string, unknown>).create_pod = mockCreatePod;

  // userCredentialService creates new objects each call — override the getter
  vi.mocked(getUserCredentialService).mockReturnValue({
    list_agent_credentials_for_agent: mockListCredentials,
  } as unknown as ReturnType<typeof getUserCredentialService>);
}

describe("useCreatePodForm - credential via agentfile_layer (SSOT)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setupMocks();
    mockListCredentials.mockResolvedValue(
      JSON.stringify({ profiles: [], runner_host: { available: true } })
    );
  });

  it("should omit CREDENTIAL from agentfile_layer when RunnerHost is selected", async () => {
    mockCreatePod.mockResolvedValue(
      JSON.stringify({ pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } })
    );

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    act(() => {
      result.current.setSelectedAgent("claude-code");
    });

    expect(result.current.selectedCredentialProfile).toBe(RUNNER_HOST_PROFILE_ID);
    expect(result.current.selectedCredentialProfile).toBe(0);

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).toHaveBeenCalledTimes(1);
    const createArg = JSON.parse(mockCreatePod.mock.calls[0][0]);
    expect(createArg).not.toHaveProperty("credential_profile_id");
    const layer = createArg.agentfile_layer ?? "";
    expect(layer).not.toContain("CREDENTIAL");
  });

  it("should include CREDENTIAL in agentfile_layer when custom profile selected", async () => {
    const customProfile = { id: 42, name: "My API Key", is_default: false, is_active: true };
    mockListCredentials.mockResolvedValue(
      JSON.stringify({ profiles: [customProfile], runner_host: { available: true } })
    );
    mockCreatePod.mockResolvedValue(
      JSON.stringify({ pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } })
    );

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    act(() => {
      result.current.setSelectedAgent("claude-code");
    });
    await act(async () => {});

    act(() => {
      result.current.setSelectedCredentialProfile(42);
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    expect(mockCreatePod).toHaveBeenCalledTimes(1);
    const createArg = JSON.parse(mockCreatePod.mock.calls[0][0]);
    expect(createArg).not.toHaveProperty("credential_profile_id");
    expect(createArg.agentfile_layer).toContain('CREDENTIAL "My API Key"');
  });

  it("should always send agentfile_layer via API (SSOT)", async () => {
    mockCreatePod.mockResolvedValue(
      JSON.stringify({ pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } })
    );

    const { result } = renderHook(() => useCreatePodForm(mockAgents, []));

    act(() => {
      result.current.setSelectedAgent("claude-code");
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const createArg = JSON.parse(mockCreatePod.mock.calls[0][0]);
    expect(createArg).toHaveProperty("agent_slug", "claude-code");
    expect(createArg).not.toHaveProperty("credential_profile_id");
    expect(createArg).not.toHaveProperty("repository_id");
    expect(createArg).not.toHaveProperty("interaction_mode");
    expect(createArg).not.toHaveProperty("branch_name");
    expect(createArg).not.toHaveProperty("prompt");
    expect(createArg).not.toHaveProperty("config_overrides");
  });

  it("should omit Factory ACP-only config from PTY agentfile_layer", async () => {
    mockCreatePod.mockResolvedValue(
      JSON.stringify({ pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } })
    );

    const { result } = renderHook(() =>
      useCreatePodForm([factoryAgent], [], undefined, { skip_permissions_unsafe: true })
    );

    act(() => {
      result.current.setSelectedAgent("factory-cli");
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const createArg = JSON.parse(mockCreatePod.mock.calls[0][0]);
    expect(createArg.agentfile_layer ?? "").not.toContain("skip_permissions_unsafe");
  });

  it("should include Factory ACP-only config in ACP agentfile_layer", async () => {
    mockCreatePod.mockResolvedValue(
      JSON.stringify({ pod: { pod_key: "test-pod", id: 1, status: "initializing", agent_status: "idle" } })
    );

    const { result } = renderHook(() =>
      useCreatePodForm([factoryAgent], [], undefined, { skip_permissions_unsafe: true })
    );

    act(() => {
      result.current.setSelectedAgent("factory-cli");
      result.current.setInteractionMode(POD_MODE_ACP);
    });

    await act(async () => {
      await result.current.submit(1, {}, { cols: 80, rows: 24 });
    });

    const createArg = JSON.parse(mockCreatePod.mock.calls[0][0]);
    expect(createArg.agentfile_layer).toContain("MODE acp");
    expect(createArg.agentfile_layer).toContain("CONFIG skip_permissions_unsafe = true");
  });

  it("should refresh raw agentfile layer from the current preview when source mode is enabled", async () => {
    const { result, rerender } = renderHook(
      ({ configValues }) => useCreatePodForm([factoryAgent], [], undefined, configValues),
      { initialProps: { configValues: { skip_permissions_unsafe: false } } }
    );

    act(() => {
      result.current.setSelectedAgent("factory-cli");
      result.current.setInteractionMode(POD_MODE_ACP);
    });
    await act(async () => {});

    act(() => {
      result.current.setRawLayerMode(true);
    });
    expect(result.current.rawLayerText).toContain("CONFIG skip_permissions_unsafe = false");

    act(() => {
      result.current.setRawLayerMode(false);
    });

    act(() => {
      rerender({ configValues: { skip_permissions_unsafe: true } });
    });
    expect(result.current.agentfileLayer).toContain("CONFIG skip_permissions_unsafe = true");

    act(() => {
      result.current.setRawLayerMode(true);
    });

    expect(result.current.rawLayerText).toContain("CONFIG skip_permissions_unsafe = true");
    expect(result.current.rawLayerText).not.toContain("CONFIG skip_permissions_unsafe = false");
  });
});
