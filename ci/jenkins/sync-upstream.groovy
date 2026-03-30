#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for Syncing Upstream Repository
 *
 * This pipeline:
 * 1. Fetches the latest changes from upstream remote
 * 2. Merges upstream/main into local main branch
 * 3. Pushes the merged changes to origin remote
 *
 * Prerequisites:
 * - Git repository must have 'upstream' remote configured
 * - Jenkins SSH credentials configured for Git authentication
 *
 * Environment Variables:
 * - GIT_REPO: Git repository URL (origin)
 * - GIT_UPSTREAM_URL: Upstream repository URL (default: https://github.com/AgentMesh/AgentsMesh.git)
 */

String getBuildUser() {
  def cause = currentBuild.getBuildCauses('hudson.model.Cause$UserIdCause')
  if (cause && cause.userId) {
    return cause.userId[0]
  }
  return "unknown"
}

currentBuild.setDescription("triggered by " + getBuildUser())

pipeline {
    agent {
        node {
            label "${params.NODE}"
        }
    }

    parameters {
        string(
            name: 'NODE',
            defaultValue: 'aqa01-i01-jpt44.int.rclabenv.com',
            description: 'Jenkins node label to run the pipeline'
        )
        string(
            name: 'UPSTREAM_URL',
            defaultValue: 'https://github.com/AgentsMesh/AgentsMesh.git',
            description: 'Upstream repository URL'
        )
        string(
            name: 'UPSTREAM_BRANCH',
            defaultValue: 'main',
            description: 'Upstream branch to sync from'
        )
        string(
            name: 'LOCAL_BRANCH',
            defaultValue: 'main',
            description: 'Local branch to sync to'
        )
        string(
            name: 'GIT_CREDENTIAL_ID',
            defaultValue: 'gitjenkins.xiamen',
            description: 'Jenkins SSH credential ID for Git push'
        )
    }

    environment {
        // Git configuration
        GIT_REPO = 'git@git.ringcentral.com:ai-testing/AgentsMesh.git'
        GIT_UPSTREAM_URL = "${params.UPSTREAM_URL}"
        GIT_UPSTREAM_BRANCH = "${params.UPSTREAM_BRANCH}"
        GIT_LOCAL_BRANCH = "${params.LOCAL_BRANCH}"
    }

    stages {
        stage('Verify Git Configuration') {
            steps {
                script {
                    echo "=== Verifying Git Configuration ==="
                    echo "Origin: ${GIT_REPO}"
                    echo "Upstream: ${GIT_UPSTREAM_URL}"
                    echo "Upstream Branch: ${GIT_UPSTREAM_BRANCH}"
                    echo "Local Branch: ${GIT_LOCAL_BRANCH}"

                    // Display current remotes
                    sh '''
                        echo "Current Git user:"
                        git config user.name
                        echo "Current Git remotes:"
                        git remote -v || true
                    '''
                }
            }
        }

        stage('Configure Upstream Remote') {
            steps {
                script {
                    echo "=== Configuring Upstream Remote ==="
                    sh """
                        # Remove existing upstream remote if exists
                        git remote remove upstream || true

                        # Add upstream remote
                        git remote add upstream ${GIT_UPSTREAM_URL}

                        # Verify remotes
                        echo "Updated Git remotes:"
                        git remote -v
                    """
                }
            }
        }

        stage('Fetch Upstream') {
            steps {
                script {
                    echo "=== Fetching from Upstream ==="
                    sh """
                        # For public GitHub repos, we need to disable credential helper
                        # to prevent authentication prompts
                        export GIT_TERMINAL_PROMPT=0
                        export GIT_ASKPASS=true

                        # Temporarily disable credential helper for this fetch
                        git config --global credential.helper ""

                        # Fetch from upstream
                        git fetch upstream || {
                            echo "❌ Direct fetch failed, trying with explicit HTTPS URL..."
                            # If upstream remote is configured with git@github.com,
                            # fetch directly from HTTPS URL instead
                            git fetch ${GIT_UPSTREAM_URL}
                        }

                        echo "✅ Successfully fetched from upstream"
                    """
                }
            }
        }

        stage('Checkout Local Branch') {
            steps {
                script {
                    echo "=== Checking out local ${GIT_LOCAL_BRANCH} branch ==="
                    sh """
                        # Ensure we're on the local branch, not detached HEAD
                        git checkout -B ${GIT_LOCAL_BRANCH} origin/${GIT_LOCAL_BRANCH}

                        # Verify we're on the correct branch
                        git branch --show-current
                        git status
                    """
                }
            }
        }

        stage('Merge Upstream') {
            steps {
                script {
                    echo "=== Merging upstream/${GIT_UPSTREAM_BRANCH} into ${GIT_LOCAL_BRANCH} ==="
                    sh """
                        git merge upstream/${GIT_UPSTREAM_BRANCH} --no-edit
                        echo "✅ Successfully merged upstream/${GIT_UPSTREAM_BRANCH}"
                    """
                }
            }
        }

        stage('Push to Origin') {
            steps {
                script {
                    echo "=== Pushing ${GIT_LOCAL_BRANCH} to origin ==="
                    // Use sshagent to provide SSH credentials for git push
                    sshagent(credentials: ["${params.GIT_CREDENTIAL_ID}"]) {
                        sh """
                            # Verify SSH key is loaded
                            ssh-add -l || echo "Warning: No SSH keys found"

                            # Ensure origin remote uses SSH URL
                            git remote set-url origin ${GIT_REPO}

                            # Push to origin
                            git push origin ${GIT_LOCAL_BRANCH}
                            echo "✅ Successfully pushed to origin/${GIT_LOCAL_BRANCH}"
                        """
                    }
                }
            }
        }
    }

    post {
        success {
            echo """
=== ✅ Sync Successful ===
Successfully synced upstream/${GIT_UPSTREAM_BRANCH} to origin/${GIT_LOCAL_BRANCH}

Details:
  - Upstream: ${GIT_UPSTREAM_URL}
  - Origin: ${GIT_REPO}
  - Branch: ${GIT_LOCAL_BRANCH}
"""
        }
        failure {
            echo """
=== ❌ Sync Failed ===
Failed to sync upstream repository.

Common issues:
  1. Merge conflicts - resolve manually and retry
  2. Network issues - check connectivity to upstream
  3. Permission issues - verify Git credentials

Details:
  - Upstream: ${GIT_UPSTREAM_URL}
  - Origin: ${GIT_REPO}
  - Branch: ${GIT_LOCAL_BRANCH}
"""
        }
        always {
            script {
                echo "=== Cleaning up ==="
                sh """
                    # Show final status
                    git status || true

                    # Clean up upstream remote (optional)
                    # git remote remove upstream || true
                """
            }
        }
    }
}
