import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:agentmesh/models/pod.dart';
import 'package:agentmesh/providers/auth_provider.dart';

class PodListState {
  final List<Pod> pods;
  final bool isLoading;
  final bool hasMore;
  final String? error;
  final String? statusFilter;

  const PodListState({
    this.pods = const [],
    this.isLoading = false,
    this.hasMore = true,
    this.error,
    this.statusFilter,
  });

  PodListState copyWith({
    List<Pod>? pods,
    bool? isLoading,
    bool? hasMore,
    String? error,
    String? statusFilter,
  }) {
    return PodListState(
      pods: pods ?? this.pods,
      isLoading: isLoading ?? this.isLoading,
      hasMore: hasMore ?? this.hasMore,
      error: error,
      statusFilter: statusFilter ?? this.statusFilter,
    );
  }

  List<Pod> get activePods =>
      pods.where((p) => p.isActive).toList();

  List<Pod> get completedPods =>
      pods.where((p) => p.isTerminated).toList();
}

class PodListNotifier extends StateNotifier<PodListState> {
  final Ref _ref;
  static const int _pageSize = 20;

  PodListNotifier(this._ref) : super(const PodListState());

  Future<void> loadPods({bool refresh = false}) async {
    if (state.isLoading) return;

    if (refresh) {
      state = state.copyWith(pods: [], hasMore: true);
    }

    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = _ref.read(apiClientProvider);
      final offset = refresh ? 0 : state.pods.length;

      final response = await apiClient.getPods(
        status: state.statusFilter,
        limit: _pageSize,
        offset: offset,
      );

      final List<dynamic> podList = response.data['pods'] ?? [];
      final newPods = podList.map((p) => Pod.fromJson(p)).toList();

      state = state.copyWith(
        pods: refresh ? newPods : [...state.pods, ...newPods],
        isLoading: false,
        hasMore: newPods.length >= _pageSize,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to load pods',
      );
    }
  }

  Future<void> refresh() => loadPods(refresh: true);

  void setStatusFilter(String? status) {
    state = state.copyWith(statusFilter: status);
    loadPods(refresh: true);
  }
}

final podListProvider =
    StateNotifierProvider<PodListNotifier, PodListState>((ref) {
  return PodListNotifier(ref);
});

class PodDetailState {
  final Pod? pod;
  final bool isLoading;
  final String? error;

  const PodDetailState({
    this.pod,
    this.isLoading = false,
    this.error,
  });

  PodDetailState copyWith({
    Pod? pod,
    bool? isLoading,
    String? error,
  }) {
    return PodDetailState(
      pod: pod ?? this.pod,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class PodDetailNotifier extends StateNotifier<PodDetailState> {
  final Ref _ref;
  final String podKey;

  PodDetailNotifier(this._ref, this.podKey)
      : super(const PodDetailState()) {
    loadPod();
  }

  Future<void> loadPod() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = _ref.read(apiClientProvider);
      final response = await apiClient.getPod(podKey);
      final pod = Pod.fromJson(response.data['pod']);

      state = state.copyWith(
        pod: pod,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to load pod',
      );
    }
  }

  Future<bool> terminatePod() async {
    try {
      final apiClient = _ref.read(apiClientProvider);
      await apiClient.terminatePod(podKey);
      await loadPod();
      return true;
    } catch (e) {
      return false;
    }
  }

  void updateFromWebSocket(Pod pod) {
    if (pod.podKey == podKey) {
      state = state.copyWith(pod: pod);
    }
  }
}

final podDetailProvider = StateNotifierProvider.family<
    PodDetailNotifier, PodDetailState, String>((ref, podKey) {
  return PodDetailNotifier(ref, podKey);
});

// Runners provider
class RunnerListState {
  final List<Runner> runners;
  final bool isLoading;
  final String? error;

  const RunnerListState({
    this.runners = const [],
    this.isLoading = false,
    this.error,
  });

  RunnerListState copyWith({
    List<Runner>? runners,
    bool? isLoading,
    String? error,
  }) {
    return RunnerListState(
      runners: runners ?? this.runners,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }

  List<Runner> get onlineRunners => runners.where((r) => r.isOnline).toList();
  List<Runner> get availableRunners =>
      runners.where((r) => r.isOnline && r.hasCapacity).toList();
}

class RunnerListNotifier extends StateNotifier<RunnerListState> {
  final Ref _ref;

  RunnerListNotifier(this._ref) : super(const RunnerListState());

  Future<void> loadRunners() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = _ref.read(apiClientProvider);
      final response = await apiClient.getRunners();

      final List<dynamic> runnerList = response.data['runners'] ?? [];
      final runners = runnerList.map((r) => Runner.fromJson(r)).toList();

      state = state.copyWith(
        runners: runners,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to load runners',
      );
    }
  }

  Future<void> refresh() => loadRunners();
}

final runnerListProvider =
    StateNotifierProvider<RunnerListNotifier, RunnerListState>((ref) {
  return RunnerListNotifier(ref);
});

// Agent types provider
class AgentTypeListState {
  final List<AgentType> types;
  final bool isLoading;
  final String? error;

  const AgentTypeListState({
    this.types = const [],
    this.isLoading = false,
    this.error,
  });

  AgentTypeListState copyWith({
    List<AgentType>? types,
    bool? isLoading,
    String? error,
  }) {
    return AgentTypeListState(
      types: types ?? this.types,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class AgentTypeListNotifier extends StateNotifier<AgentTypeListState> {
  final Ref _ref;

  AgentTypeListNotifier(this._ref) : super(const AgentTypeListState());

  Future<void> loadAgentTypes() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = _ref.read(apiClientProvider);
      final response = await apiClient.getAgentTypes();

      final List<dynamic> typeList = response.data['agent_types'] ?? [];
      final types = typeList.map((t) => AgentType.fromJson(t)).toList();

      state = state.copyWith(
        types: types,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Failed to load agent types',
      );
    }
  }
}

final agentTypeListProvider =
    StateNotifierProvider<AgentTypeListNotifier, AgentTypeListState>((ref) {
  return AgentTypeListNotifier(ref);
});
