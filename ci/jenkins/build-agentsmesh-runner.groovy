#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for AgentsMesh Runner Binary Build
 *
 * This pipeline:
 * 1. Clones the repository from GitLab
 * 2. Builds Runner binaries for all platforms (macOS, Linux, Windows, amd64/arm64)
 * 3. Uploads built binaries to MinIO
 * 4. Archives built binaries as Jenkins artifacts
 *
 * Build Command:
 * - cd runner && make build-all
 *
 * Build Artifacts:
 * - runner/bin/runner-darwin-amd64
 * - runner/bin/runner-darwin-arm64
 * - runner/bin/runner-linux-amd64
 * - runner/bin/runner-linux-arm64
 * - runner/bin/runner-windows-amd64.exe
 * - runner/bin/runner-windows-arm64.exe
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

        // Build metadata
        VERSION = "dev" // sh(script: "git describe --tags --always --dirty 2>/dev/null || echo 'dev'", returnStdout: true).trim()
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

        stage('Build Runner') {
            steps {
                script {
                    echo "=== Building AgentsMesh Runner for all platforms ==="
                    echo "Version: ${VERSION}"
                    echo "Build Time: ${BUILD_TIME}"

                    sh """
                        source ~/.bashrc
                        cd runner
                        make build-all VERSION=${VERSION} BUILD_TIME=${BUILD_TIME}
                    """

                    echo "=== Build complete ==="
                    sh "ls -lh runner/bin/"
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

                    def artifacts = [
                        "runner/bin/runner-darwin-amd64",
                        "runner/bin/runner-darwin-arm64",
                        "runner/bin/runner-linux-amd64",
                        "runner/bin/runner-linux-arm64",
                        "runner/bin/runner-windows-amd64.exe",
                        "runner/bin/runner-windows-arm64.exe"
                    ]

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
                        def fileName = artifact.split('/').last()
                        echo "Uploading ${fileName}..."

                        sh """
                            # Ensure mcli is in PATH
                            export PATH=\$HOME/.local/bin:\$PATH

                            mcli cp "${artifact}" agentsmesh-minio/${MINIO_BUCKET}/${fileName}
                        """

                        // Construct public download URL
                        def downloadUrl = "${MINIO_ENDPOINT}/${MINIO_BUCKET}/${fileName}"
                        minioLinks.add([name: fileName, url: downloadUrl])
                        echo "Uploaded ${fileName}: ${downloadUrl}"
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
                archiveArtifacts artifacts: 'runner/bin/runner-*', fingerprint: true

                def artifacts = [
                    "runner/bin/runner-darwin-amd64",
                    "runner/bin/runner-darwin-arm64",
                    "runner/bin/runner-linux-amd64",
                    "runner/bin/runner-linux-arm64",
                    "runner/bin/runner-windows-amd64.exe",
                    "runner/bin/runner-windows-arm64.exe"
                ]

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
