package runner

// GetRecentOutput returns recent terminal output for observation (plain text without ANSI)
func (tr *TerminalRouter) GetRecentOutput(podKey string, lines int) []byte {
	shard := tr.getShard(podKey)

	shard.mu.RLock()
	vt := shard.virtualTerminals[podKey]
	shard.mu.RUnlock()

	if vt == nil {
		return nil
	}

	output := vt.GetOutput(lines)
	if output == "" {
		return nil
	}
	return []byte(output)
}

// GetScreenSnapshot returns the current screen snapshot for agent observation
func (tr *TerminalRouter) GetScreenSnapshot(podKey string) string {
	shard := tr.getShard(podKey)

	shard.mu.RLock()
	vt := shard.virtualTerminals[podKey]
	shard.mu.RUnlock()

	if vt == nil {
		return ""
	}
	return vt.GetDisplay()
}

// GetCursorPosition returns the current cursor position (row, col) for a pod
func (tr *TerminalRouter) GetCursorPosition(podKey string) (row, col int) {
	shard := tr.getShard(podKey)

	shard.mu.RLock()
	vt := shard.virtualTerminals[podKey]
	shard.mu.RUnlock()

	if vt == nil {
		return 0, 0
	}
	return vt.CursorPosition()
}

// ClearTerminal clears the virtual terminal state for a pod
func (tr *TerminalRouter) ClearTerminal(podKey string) {
	shard := tr.getShard(podKey)

	shard.mu.RLock()
	vt := shard.virtualTerminals[podKey]
	shard.mu.RUnlock()

	if vt != nil {
		vt.Clear()
	}
}
