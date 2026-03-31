FROM node:24-bookworm-slim

WORKDIR /app
ENV HOME=/home/node

# ── System dependencies ────────────────────────────────────────
RUN apt-get update && apt-get install -y \
    ca-certificates \
    build-essential \
    python3 python3-pip python3-dev python-is-python3 \
    git curl bash netcat-openbsd \
    openssh-client \
    chromium \
    fonts-noto-cjk \
    fonts-freefont-ttf \
    libnss3 \
    libfreetype6 \
    libharfbuzz0b \
    libvulkan1 \
    wget \
    && rm -rf /var/lib/apt/lists/*

# ── Install Google Chrome ──────────────────────────────────────
RUN curl -fsSL https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb -o /tmp/chrome.deb \
    && apt-get install -y /tmp/chrome.deb \
    && rm /tmp/chrome.deb \
    && rm -rf /var/lib/apt/lists/*

# ============================================
# Install AI CLI Tools (as root, before user switch)
# ============================================

# 1. Claude Code - Anthropic's AI coding assistant
RUN npm install -g @anthropic-ai/claude-code

# 2. OpenAI Codex CLI - OpenAI's coding agent
RUN npm install -g @openai/codex

# 3. Gemini CLI - Google's AI coding assistant
RUN npm install -g @google/gemini-cli

# 4. OpenCode - Open source AI coding agent
RUN npm install -g opencode-ai

# 5. Auggie - AugmentCode's AI coding agent
RUN npm install -g @augmentcode/auggie

# 6. Cursor - Cursor's AI coding assistant
RUN curl -fsSL https://cursor.sh/install | sh

# 7. Factory - Factory's AI coding agent
RUN export HOME=/usr/local && \
    curl -fsSL https://app.factory.ai/cli | sh

# Ensure PATH includes installers' default locations
ENV PATH="/usr/local/.local/bin:/root/.local/bin:/root/.factory/bin:${PATH}"

# Verify installations
RUN echo "=== Verifying AI CLI installations ===" && \
    claude --version && \
    codex --version && \
    gemini --version && \
    which opencode && echo "OpenCode installed at $(which opencode)" && \
    echo "=== All AI CLI tools installed ==="

# ============================================
# Create non-root user for security
# ============================================

# Reuse existing non-root user from base image (node: UID/GID 1000)
RUN set -eux; \
    mkdir -p /workspace /tmp/agentsmesh-workspace; \
    mkdir -p /home/node/.agentsmesh /home/node/.cache/go-build; \
    mkdir -p /home/node/.claude /home/node/.codex /home/node/.gemini /home/node/.opencode /home/node/.cursor /home/node/.auggie /home/node/.factory; \
    chown -R node:node /app /workspace /tmp/agentsmesh-workspace /home/node

# ============================================
# Copy AI CLI pre-configured settings
# These settings enable headless/non-interactive mode
# ============================================
# COPY --chown=runner:runner deploy/dev/ai-cli-configs/claude/settings.json /home/runner/.claude/settings.json
# COPY --chown=runner:runner deploy/dev/ai-cli-configs/codex/config.toml /home/runner/.codex/config.toml
# COPY --chown=runner:runner deploy/dev/ai-cli-configs/gemini/settings.json /home/runner/.gemini/settings.json

USER node
WORKDIR /workspace

# ── Install Playwright Chromium (as node user) ─────────────────
RUN npx playwright install chromium

# Keep container running for exec-based CLI usage
CMD ["sleep", "infinity"]