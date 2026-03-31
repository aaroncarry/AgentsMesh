#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for AgentsMesh Docker Image Build & Push
 *
 * This pipeline:
 * 1. Clones the repository from GitLab
 * 2. Builds Docker images for backend, relay, web, web-admin, and rc-runner (selectable)
 * 3. Pushes images to Harbor registry (harbor-xmn.int.rclabenv.com/agentsmesh)
 *
 * Required Jenkins Credentials:
 * - 'harbor-credentials': Username/Password credential for Harbor registry
 *
 * Parameters:
 * - NODE: Jenkins node label to run the pipeline (default: aqa01-i01-jpt44.int.rclabenv.com)
 * - BRANCH: Git branch to build (default: main)
 * - BUILD_BACKEND: Whether to build Backend image (default: true)
 * - BUILD_RELAY: Whether to build Relay image (default: true)
 * - BUILD_WEB: Whether to build Web image (default: true)
 * - BUILD_WEB_ADMIN: Whether to build Web-Admin image (default: true)
 * - BUILD_RC_RUNNER: Whether to build RC-Runner image (default: true)
 *
 * Environment Variables:
 * - HARBOR_REGISTRY: Harbor registry URL (default: harbor-xmn.int.rclabenv.com)
 * - HARBOR_PROJECT: Harbor project name (default: agentsmesh)
 * - GIT_REPO: Git repository URL
 * - GIT_BRANCH: Git branch to build (from BRANCH parameter)
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
    
    environment {
        // Registry configuration
        HARBOR_REGISTRY = 'harbor-xmn.int.rclabenv.com'
        HARBOR_PROJECT = 'agentsmesh'
        
        // Git configuration
        GIT_REPO = 'git@git.ringcentral.com:ai-testing/AgentsMesh.git'
        GIT_BRANCH = "${params.BRANCH}"
        
        // Image tags
        GIT_COMMIT_SHORT = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
        IMAGE_TAG = "${env.BUILD_NUMBER}-${GIT_COMMIT_SHORT}"
    }

    parameters {
        string(
            name: 'NODE',
            defaultValue: 'aqa01-i01-jpt44.int.rclabenv.com',
            description: 'Jenkins node label to run the pipeline'
        )
        string(
            name: 'BRANCH',
            defaultValue: 'main',
            description: 'Git branch to build'
        )
        booleanParam(
            name: 'BUILD_BACKEND',
            defaultValue: true,
            description: 'Build Backend image'
        )
        booleanParam(
            name: 'BUILD_RELAY',
            defaultValue: true,
            description: 'Build Relay image'
        )
        booleanParam(
            name: 'BUILD_WEB',
            defaultValue: true,
            description: 'Build Web image'
        )
        booleanParam(
            name: 'BUILD_WEB_ADMIN',
            defaultValue: true,
            description: 'Build Web-Admin image'
        )
        booleanParam(
            name: 'BUILD_RC_RUNNER',
            defaultValue: true,
            description: 'Build RC-Runner image'
        )
    }
    
    stages {
        stage('Docker Login') {
            steps {
                script {
                    echo "=== Logging in to Harbor registry ==="
                    withCredentials([usernamePassword(
                        credentialsId: 'harbor-credentials',
                        usernameVariable: 'HARBOR_USER',
                        passwordVariable: 'HARBOR_PASS'
                    )]) {
                        sh """
                            echo "\${HARBOR_PASS}" | docker login -u "\${HARBOR_USER}" --password-stdin ${HARBOR_REGISTRY}
                        """
                    }
                }
            }
        }
        
        stage('Build & Push Backend') {
            when {
                expression { params.BUILD_BACKEND == true }
            }
            steps {
                script {
                    echo "=== Building Backend image ==="
                    buildAndPushImage('backend', 'ci/backend.Dockerfile')
                }
            }
        }

        stage('Build & Push Relay') {
            when {
                expression { params.BUILD_RELAY == true }
            }
            steps {
                script {
                    echo "=== Building Relay image ==="
                    buildAndPushImage('relay', 'ci/relay.Dockerfile')
                }
            }
        }

        stage('Build & Push Web') {
            when {
                expression { params.BUILD_WEB == true }
            }
            steps {
                script {
                    echo "=== Building Web image ==="
                    buildAndPushImage('web', 'ci/web.Dockerfile')
                }
            }
        }

        stage('Build & Push Web-Admin') {
            when {
                expression { params.BUILD_WEB_ADMIN == true }
            }
            steps {
                script {
                    echo "=== Building Web-Admin image ==="
                    buildAndPushImage('web-admin', 'ci/web-admin.Dockerfile')
                }
            }
        }

        stage('Build & Push RC-Runner') {
            when {
                expression { params.BUILD_RC_RUNNER == true }
            }
            steps {
                script {
                    echo "=== Building RC-Runner image ==="
                    buildAndPushImage('rc-runner', 'ci/rc-runner.Dockerfile')
                }
            }
        }
    }
    
    post {
        success {
            script {
                def builtImages = []
                if (params.BUILD_BACKEND) {
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/backend:${IMAGE_TAG}"
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/backend:latest"
                }
                if (params.BUILD_RELAY) {
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/relay:${IMAGE_TAG}"
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/relay:latest"
                }
                if (params.BUILD_WEB) {
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/web:${IMAGE_TAG}"
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/web:latest"
                }
                if (params.BUILD_WEB_ADMIN) {
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/web-admin:${IMAGE_TAG}"
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/web-admin:latest"
                }
                if (params.BUILD_RC_RUNNER) {
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/rc-runner:${IMAGE_TAG}"
                    builtImages << "  - ${HARBOR_REGISTRY}/${HARBOR_PROJECT}/rc-runner:latest"
                }

                def message = """
=== ✅ Build Successful ===
The following images have been built and pushed to Harbor:
${builtImages.join('\n')}
"""
                echo message
            }
        }
        failure {
            echo "=== ❌ Build Failed ==="
        }
        always {
            script {
                echo "=== Cleaning up ==="
                sh "docker logout ${HARBOR_REGISTRY} || true"
            }
        }
    }
}

/**
 * Build and push Docker image
 * @param component Component name (backend, relay, web, web-admin, rc-runner)
 * @param dockerfile Dockerfile path relative to project root
 */
def buildAndPushImage(String component, String dockerfile) {
    def imageName = "${HARBOR_REGISTRY}/${HARBOR_PROJECT}/${component}"
    def buildContext = '.'

    try {
        // Build image with version tag
        echo "Building ${component} image..."
        sh """
            docker build \
                -f ${dockerfile} \
                -t ${imageName}:${IMAGE_TAG} \
                ${buildContext}
        """

        // Tag with 'latest'
        echo "Tagging ${component} image as latest..."
        sh "docker tag ${imageName}:${IMAGE_TAG} ${imageName}:latest"

        // Push version tag
        echo "Pushing ${component}:${IMAGE_TAG} to Harbor..."
        sh "docker push ${imageName}:${IMAGE_TAG}"

        // Push latest tag
        echo "Pushing ${component}:latest to Harbor..."
        sh "docker push ${imageName}:latest"

        echo "✅ ${component} image successfully built and pushed"

    } catch (Exception e) {
        echo "❌ Failed to build/push ${component} image: ${e.message}"
        throw e
    } finally {
        // Clean up local images to save disk space
        sh """
            docker rmi ${imageName}:${IMAGE_TAG} || true
            docker rmi ${imageName}:latest || true
        """
    }
}