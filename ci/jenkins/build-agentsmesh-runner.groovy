#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for AgentsMesh Runner Binary Build
 *
 * This pipeline:
 * 1. Clones the repository from GitLab
 * 2. Builds Runner binaries and packages for all platforms using GoReleaser
 * 3. Uploads built artifacts to MinIO
 * 4. Archives built artifacts as Jenkins artifacts
 *
 * Build Command:
 * - cd runner && goreleaser release --snapshot --clean
 *
 * Build Artifacts (dist/):
 * - agentsmesh-runner_darwin_all.tar.gz       (macOS universal binary)
 * - agentsmesh-runner_darwin_amd64.tar.gz     (macOS Intel)
 * - agentsmesh-runner_darwin_arm64.tar.gz     (macOS Apple Silicon)
 * - agentsmesh-runner_linux_amd64.tar.gz      (Linux amd64)
 * - agentsmesh-runner_linux_arm64.tar.gz      (Linux arm64)
 * - agentsmesh-runner_linux_amd64.deb         (Debian/Ubuntu amd64)
 * - agentsmesh-runner_linux_arm64.deb         (Debian/Ubuntu arm64)
 * - agentsmesh-runner_linux_amd64.rpm         (RedHat/CentOS amd64)
 * - agentsmesh-runner_linux_arm64.rpm         (RedHat/CentOS arm64)
 * - agentsmesh-runner_linux_amd64.apk         (Alpine amd64)
 * - agentsmesh-runner_linux_arm64.apk         (Alpine arm64)
 * - agentsmesh-runner_windows_amd64.zip       (Windows amd64)
 * - agentsmesh-runner_windows_arm64.zip       (Windows arm64)
 *
 * MinIO Upload:
 * - Endpoint: http://aqa01-i01-xta01.int.rclabenv.com:9900
 * - Bucket: agentsmesh
 * - MinIO Client will be automatically installed as 'mcli' if not present (to avoid conflict with GNU Midnight Commander)
 * - All artifacts are uploaded directly (overwrites existing files)
 * - Files are publicly accessible
 *
 * Parameters:
 * - NODE: Jenkins node label to run the pipeline (default: aqa01-i01-jpt44.int.rclabenv.com)
 * - BRANCH: Git branch to build (default: rc)
 * - GIT_CREDENTIAL_ID: Jenkins SSH credential ID for Git operations (default: gitjenkins.xiamen)
 * - UPLOAD_TO_MINIO: Upload artifacts to MinIO (default: true)
 *
 * Environment Variables:
 * - GIT_REPO: Git repository URL
 * - GIT_BRANCH: Git branch to build (from BRANCH parameter)
 * - VERSION: Git version (tag or commit hash)
 * - BUILD_TIME: Build timestamp
 * - MINIO_ENDPOINT: MinIO server endpoint
 * - MINIO_BUCKET: MinIO bucket name
 * - MINIO_ACCESS_KEY: MinIO access key
 * - MINIO_SECRET_KEY: MinIO secret key
 */

String getBuildUser() {
  def cause = currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause')
  if (cause && cause.userId) {
    return cause.userId[0]
  }
  return "unknown"
}

currentBuild.setDescription("triggered by " + getBuildUser())

// Global variable to store public download links
def minioLinks = []

pipeline {
    agent {
        node {
            label "${params.NODE}"
        }
    }

    environment {
        // Git configuration
        GIT_REPO = 'git@git.ringcentral.com:ai-testing/AgentsMesh.git'
        GIT_BRANCH = "${params.BRANCH}"

        // Runner version
        VERSION = ''

        // Build metadata (VERSION will be set dynamically in stages)
        BUILD_TIME = sh(script: "date -u '+%Y-%m-%d_%H:%M:%S'", returnStdout: true).trim()

        // MinIO configuration
        MINIO_ENDPOINT = 'http://aqa01-i01-xta01.int.rclabenv.com:9900'
        MINIO_BUCKET = 'agentsmesh'
        MINIO_ACCESS_KEY = 'admin'
        MINIO_SECRET_KEY = '12345678'
    }

    parameters {
        string(
            name: 'NODE',
            defaultValue: 'aqa01-i01-jpt44.int.rclabenv.com',
            description: 'Jenkins node label to run the pipeline'
        )
        string(
            name: 'BRANCH',
            defaultValue: 'rc',
            description: 'Git branch to build'
        )
        string(
            name: 'GIT_CREDENTIAL_ID',
            defaultValue: 'gitjenkins.xiamen',
            description: 'Jenkins SSH credential ID for Git operations'
        )
        password(
            name: 'GITHUB_PAT',
            defaultValue: 'github-pat-for-agentsmesh-build-runner',
            description: 'Jenkins credential ID for GitHub API token (used to fetch latest release version)'
        )
        booleanParam(
            name: 'UPLOAD_TO_MINIO',
            defaultValue: true,
            description: 'Upload build artifacts to MinIO'
        )
    }

    stages {
        stage('Checkout Branch') {
            steps {
                script {
                    echo "=== Checking out branch: ${params.BRANCH} ==="
                    sshagent(credentials: ["${params.GIT_CREDENTIAL_ID}"]) {
                        sh """
                            # Verify SSH key is loaded
                            ssh-add -l || echo "Warning: No SSH keys found"

                            # Ensure origin remote uses SSH URL
                            git remote set-url origin ${GIT_REPO}

                            # Fetch latest changes
                            git fetch origin

                            # Checkout and pull the specified branch
                            git checkout ${params.BRANCH}
                            git pull origin ${params.BRANCH}

                            echo "Current branch: \$(git branch --show-current)"
                            echo "Current commit: \$(git rev-parse HEAD)"
                        """
                    }
                }
            }
        }

        stage('Get Runner Latest Version') {
            steps {
                script {
                    echo "=== Fetching latest version from GitHub API ==="

                    // Get version from GitHub API with authentication
                    withCredentials([string(credentialsId: "${params.GITHUB_PAT}", variable: 'GITHUB_PAT')]) {
                        VERSION = sh(script: '''
                            GITHUB_REPO="AgentsMesh/AgentsMesh"

                            # Get latest version from GitHub API with authentication
                            LATEST_VERSION=$(curl -sL \
                                -H "Authorization: token ${GITHUB_PAT}" \
                                -H "Accept: application/vnd.github+json" \
                                "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
                                | grep '"tag_name"' \
                                | sed -E 's/.*"v([^"]+)".*/\\1/' \
                                | head -n1)

                            if [ -z "$LATEST_VERSION" ]; then
                                echo "Error: Failed to fetch latest version from GitHub API" >&2
                                exit 1
                            fi

                            # Append commit info for development builds
                            COMMIT_SHORT=$(git rev-parse --short HEAD)
                            echo "${LATEST_VERSION}-dev-${COMMIT_SHORT}"
                        ''', returnStdout: true).trim()
                    }

                    echo "Version: ${env.VERSION}"
                }
            }
        }

        stage('Build Runner') {
            steps {
                script {
                    echo "=== Building AgentsMesh Runner with GoReleaser ==="
                    echo "Version: ${VERSION}"
                    echo "Build Time: ${BUILD_TIME}"

                    // Build with GoReleaser
                    sh """
                        source ~/.bashrc

                        # Set empty macOS code signing environment variables
                        # These are required by .goreleaser.yml but not used in snapshot builds
                        export MACOS_CERTIFICATE=""
                        export MACOS_CERTIFICATE_PASSWORD=""
                        export APPLE_API_KEY_ID=""
                        export APPLE_API_KEY_ISSUER_ID=""
                        export APPLE_API_KEY=""

                        # Verify goreleaser is available
                        echo "Using goreleaser: \$(which goreleaser)"
                        goreleaser --version

                        cd runner
                        goreleaser release --snapshot --clean
                    """

                    echo "=== Build complete ==="
                    sh "ls -lh runner/dist/"
                }
            }
        }

        stage('Upload to MinIO') {
            when {
                expression { params.UPLOAD_TO_MINIO == true }
            }
            steps {
                script {
                    echo "=== Uploading artifacts to MinIO ==="

                    // Collect all artifacts from dist/ directory
                    def artifacts = []

                    // Find all tar.gz, zip, deb, rpm, apk files in dist/ directory
                    sh """
                        cd runner/dist
                        find . -type f \\( -name "*.tar.gz" -o -name "*.zip" -o -name "*.deb" -o -name "*.rpm" -o -name "*.apk" \\) | sed 's|^\\./||' > /tmp/artifacts.txt
                    """

                    // Read artifact list
                    def artifactList = sh(script: "cat /tmp/artifacts.txt", returnStdout: true).trim().split('\n')

                    // Convert to full paths
                    artifactList.each { artifact ->
                        if (artifact && !artifact.contains('checksums.txt')) {
                            artifacts.add("runner/dist/${artifact}")
                        }
                    }

                    echo "Found ${artifacts.size()} artifacts to upload:"
                    artifacts.each { artifact ->
                        echo "  - ${artifact}"
                    }

                    // Check if mcli (MinIO Client) is installed, if not install it
                    sh """
                        if ! command -v mcli &> /dev/null; then
                            echo "MinIO Client (mcli) not found, installing..."

                            # Determine OS and architecture
                            OS=\$(uname -s | tr '[:upper:]' '[:lower:]')
                            ARCH=\$(uname -m)

                            # Map architecture names
                            case "\$ARCH" in
                                x86_64)
                                    ARCH="amd64"
                                    ;;
                                aarch64|arm64)
                                    ARCH="arm64"
                                    ;;
                            esac

                            # Download mc binary and rename to mcli to avoid conflict with GNU Midnight Commander
                            MC_URL="https://dl.min.io/client/mc/release/\${OS}-\${ARCH}/mc"
                            echo "Downloading from: \$MC_URL"

                            curl -o /tmp/mcli \$MC_URL
                            chmod +x /tmp/mcli

                            # Move to user's local bin (no sudo required)
                            mkdir -p \$HOME/.local/bin
                            mv /tmp/mcli \$HOME/.local/bin/mcli
                            export PATH=\$HOME/.local/bin:\$PATH

                            echo "MinIO Client installed successfully as 'mcli'"
                            mcli --version
                        else
                            echo "MinIO Client already installed as 'mcli'"
                            mcli --version
                        fi
                    """

                    // Configure MinIO client
                    sh """
                        # Ensure mcli is in PATH
                        export PATH=\$HOME/.local/bin:\$PATH

                        # Configure mcli alias (for HTTP endpoints)
                        mcli alias set agentsmesh-minio ${MINIO_ENDPOINT} ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY}

                        # Create bucket if not exists
                        mcli mb agentsmesh-minio/${MINIO_BUCKET} --ignore-existing

                        # Set bucket policy to public (download-only)
                        mcli anonymous set download agentsmesh-minio/${MINIO_BUCKET}
                    """

                    // Upload each artifact
                    artifacts.each { artifact ->
                        def originalFileName = artifact.split('/').last()

                        // Remove version from filename
                        // Format: agentsmesh-runner_VERSION_OS_ARCH.EXT -> agentsmesh-runner_OS_ARCH.EXT
                        def targetFileName = originalFileName.replaceAll(/_v?[0-9]+\.[0-9]+\.[0-9]+[^_]*_/, '_')

                        echo "Uploading ${originalFileName} -> ${targetFileName}..."

                        sh """
                            # Ensure mcli is in PATH
                            export PATH=\$HOME/.local/bin:\$PATH

                            mcli cp "${artifact}" agentsmesh-minio/${MINIO_BUCKET}/${targetFileName}
                        """

                        // Construct public download URL
                        def downloadUrl = "${MINIO_ENDPOINT}/${MINIO_BUCKET}/${targetFileName}"
                        minioLinks.add([name: targetFileName, url: downloadUrl])
                        echo "Uploaded ${targetFileName}: ${downloadUrl}"
                    }

                    echo "=== Upload complete ==="
                    echo ""
                    echo "📦 Public Download Links:"
                    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                    minioLinks.each { link ->
                        echo "  ${link.name}"
                        echo "  └─ ${link.url}"
                        echo ""
                    }
                    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                    echo "All files are publicly accessible without authentication"
                }
            }
        }
    }

    post {
        success {
            script {
                echo "=== Archiving build artifacts ==="
                // Archive all package files separately (Jenkins doesn't support brace expansion)
                archiveArtifacts artifacts: 'runner/dist/*.tar.gz, runner/dist/*.zip, runner/dist/*.deb, runner/dist/*.rpm, runner/dist/*.apk', fingerprint: true, allowEmptyArchive: true

                // Collect all artifacts for display
                def artifacts = []
                def artifactList = sh(script: "find runner/dist -type f \\( -name '*.tar.gz' -o -name '*.zip' -o -name '*.deb' -o -name '*.rpm' -o -name '*.apk' \\) -printf '%P\\n' | sort", returnStdout: true).trim().split('\n')

                artifactList.each { artifact ->
                    if (artifact && !artifact.contains('checksums.txt')) {
                        artifacts.add("runner/dist/${artifact}")
                    }
                }

                echo ""
                echo "╔════════════════════════════════════════════════════════════════════════════╗"
                echo "║                          ✅ Build Successful                               ║"
                echo "╚════════════════════════════════════════════════════════════════════════════╝"
                echo ""
                echo "Version: ${VERSION}"
                echo "Build Time: ${BUILD_TIME}"
                echo "Build Number: #${env.BUILD_NUMBER}"
                echo ""
                echo "Built Artifacts:"
                artifacts.each { artifact ->
                    echo "  ✓ ${artifact}"
                }
                echo ""

                if (params.UPLOAD_TO_MINIO && minioLinks.size() > 0) {
                    echo "╔════════════════════════════════════════════════════════════════════════════╗"
                    echo "║                      📦 MinIO Public Download Links                        ║"
                    echo "╚════════════════════════════════════════════════════════════════════════════╝"
                    echo ""
                    minioLinks.each { link ->
                        echo "📄 ${link.name}"
                        echo "   ${link.url}"
                        echo ""
                    }
                    echo "💡 These links are publicly accessible without authentication"
                    echo ""
                }

                echo "════════════════════════════════════════════════════════════════════════════"
                echo "Jenkins artifacts: ${env.BUILD_URL}artifact/"
                echo "════════════════════════════════════════════════════════════════════════════"
            }
        }
        failure {
            echo "=== ❌ Build Failed ==="
        }
        always {
            script {
                echo "=== Cleaning up ==="
            }
        }
    }
}
