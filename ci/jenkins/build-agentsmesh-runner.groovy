#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for AgentsMesh Runner Binary Build
 *
 * This pipeline:
 * 1. Clones the repository from GitLab
 * 2. Builds Runner binaries for all platforms (macOS, Linux, Windows, amd64/arm64)
 * 3. Uploads built binaries to NextCloud
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
 * NextCloud Upload:
 * - URL: https://cloud-xmn.int.rclabenv.com
 * - Directory: AgentsMesh/
 * - All artifacts are uploaded directly (overwrites existing files)
 * - Files are shared publicly (no authentication required for download)
 *
 * Parameters:
 * - NODE: Jenkins node label to run the pipeline (default: aqa01-i01-jpt44.int.rclabenv.com)
 * - BRANCH: Git branch to build (default: rc)
 * - GIT_CREDENTIAL_ID: Jenkins SSH credential ID for Git operations (default: gitjenkins.xiamen)
 * - UPLOAD_TO_NEXT_CLOUD: Upload artifacts to NextCloud (default: true)
 *
 * Environment Variables:
 * - GIT_REPO: Git repository URL
 * - GIT_BRANCH: Git branch to build (from BRANCH parameter)
 * - VERSION: Git version (tag or commit hash)
 * - BUILD_TIME: Build timestamp
 * - NEXTCLOUD_URL: NextCloud server URL
 * - NEXTCLOUD_DIR: NextCloud directory for uploads
 * - NEXTCLOUD_USER: NextCloud username
 * - NEXTCLOUD_PASS: NextCloud password
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
def nextCloudLinks = []

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

        // NextCloud configuration
        NEXTCLOUD_URL = 'https://cloud-xmn.int.rclabenv.com'
        NEXTCLOUD_DIR = 'AgentsMesh'
        NEXTCLOUD_USER = 'sdet'
        NEXTCLOUD_PASS = 'Sdet!123456'
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
            name: 'UPLOAD_TO_NEXT_CLOUD',
            defaultValue: true,
            description: 'Upload build artifacts to NextCloud'
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

        stage('Upload to NextCloud') {
            when {
                expression { params.UPLOAD_TO_NEXT_CLOUD == true }
            }
            steps {
                script {
                    echo "=== Uploading artifacts to NextCloud ==="

                    def artifacts = [
                        "runner/bin/runner-darwin-amd64",
                        "runner/bin/runner-darwin-arm64",
                        "runner/bin/runner-linux-amd64",
                        "runner/bin/runner-linux-arm64",
                        "runner/bin/runner-windows-amd64.exe",
                        "runner/bin/runner-windows-arm64.exe"
                    ]

                    // Upload each artifact and create public share link
                    artifacts.each { artifact ->
                        def fileName = artifact.split('/').last()
                        echo "Uploading ${fileName}..."

                        sh """
                            curl -k -u ${NEXTCLOUD_USER}:${NEXTCLOUD_PASS} \
                                 -T "${artifact}" \
                                 "${NEXTCLOUD_URL}/remote.php/dav/files/${NEXTCLOUD_USER}/${NEXTCLOUD_DIR}/${fileName}"
                        """

                        // Create public share link (password-free download)
                        def shareResponse = sh(
                            script: """
                                curl -k -s -u ${NEXTCLOUD_USER}:${NEXTCLOUD_PASS} \
                                     -X POST \
                                     -H "OCS-APIRequest: true" \
                                     -H "Content-Type: application/x-www-form-urlencoded" \
                                     -d "path=/${NEXTCLOUD_DIR}/${fileName}&shareType=3&permissions=1" \
                                     "${NEXTCLOUD_URL}/ocs/v2.php/apps/files_sharing/api/v1/shares?format=json"
                            """,
                            returnStdout: true
                        ).trim()

                        // Extract share URL from JSON response
                        def shareUrl = ""
                        try {
                            def jsonSlurper = new groovy.json.JsonSlurper()
                            def jsonResponse = jsonSlurper.parseText(shareResponse)
                            shareUrl = jsonResponse.ocs?.data?.url ?: "N/A"
                        } catch (Exception e) {
                            echo "Warning: Failed to parse share response for ${fileName}: ${e.message}"
                            shareUrl = "${NEXTCLOUD_URL}/apps/files/?dir=/${NEXTCLOUD_DIR}"
                        }

                        nextCloudLinks.add([name: fileName, url: shareUrl])
                        echo "Share created for ${fileName}: ${shareUrl}"
                    }

                    echo "=== Upload complete ==="
                    echo ""
                    echo "📦 Public Download Links:"
                    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
                    nextCloudLinks.each { link ->
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

                if (params.UPLOAD_TO_NEXT_CLOUD && nextCloudLinks.size() > 0) {
                    echo "╔════════════════════════════════════════════════════════════════════════════╗"
                    echo "║                     📦 NextCloud Public Download Links                     ║"
                    echo "╚════════════════════════════════════════════════════════════════════════════╝"
                    echo ""
                    nextCloudLinks.each { link ->
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
