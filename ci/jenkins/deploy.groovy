#!/usr/bin/env groovy

/**
 * Jenkins Pipeline for AgentsMesh Test Environment Deployment
 *
 * This pipeline:
 * 1. Changes to deployment directory (~/workspace/AgentsMesh/deploy/selfhost)
 * 2. Stops all AgentsMesh services (docker compose down)
 * 3. Removes old Docker images (backend, web, relay, web-admin)
 * 4. Reads DB_PASSWORD from .env file
 * 5. Starts all AgentsMesh services (docker compose up -d)
 * 6. Runs database migrations
 *
 * Prerequisites:
 * - Docker and Docker Compose installed on Jenkins node
 * - ~/workspace/AgentsMesh/deploy/selfhost/.env file configured
 *
 * Environment Variables:
 * - DEPLOY_DIR: Deployment directory (default: ~/workspace/AgentsMesh/deploy/selfhost)
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
        // Deployment directory
        DEPLOY_DIR = "${params.DEPLOY_DIR}"
    }

    parameters {
        string(
            name: 'NODE',
            defaultValue: 'aqa01-i01-jpt44.int.rclabenv.com',
            description: 'Jenkins node label to run the pipeline'
        )
        string(
            name: 'DEPLOY_DIR',
            defaultValue: '~/workspace/AgentsMesh/deploy/selfhost',
            description: 'AgentsMesh deployment directory'
        )
    }

    stages {
        stage('Verify Environment') {
            steps {
                script {
                    echo "=== Verifying deployment environment ==="
                    sh """
                        echo "Deployment directory: ${DEPLOY_DIR}"
                        echo "Checking if directory exists..."
                        if [ ! -d ${DEPLOY_DIR} ]; then
                            echo "ERROR: Deployment directory does not exist: ${DEPLOY_DIR}"
                            exit 1
                        fi

                        echo "Checking if .env file exists..."
                        if [ ! -f ${DEPLOY_DIR}/.env ]; then
                            echo "ERROR: .env file not found in ${DEPLOY_DIR}"
                            exit 1
                        fi

                        echo "Environment verification passed ✅"
                    """
                }
            }
        }

        stage('Stop Services') {
            steps {
                script {
                    echo "=== Stopping AgentsMesh services ==="
                    sh """
                        cd ${DEPLOY_DIR}
                        echo "Running: docker compose down"
                        docker compose down || true
                        echo "Services stopped ✅"
                    """
                }
            }
        }

        stage('Remove Old Images') {
            steps {
                script {
                    echo "=== Removing old Docker images ==="
                    sh """
                        echo "Removing agentsmesh/backend..."
                        docker rmi agentsmesh/backend || echo "Image agentsmesh/backend not found (skipped)"

                        echo "Removing agentsmesh/web..."
                        docker rmi agentsmesh/web || echo "Image agentsmesh/web not found (skipped)"

                        echo "Removing agentsmesh/relay..."
                        docker rmi agentsmesh/relay || echo "Image agentsmesh/relay not found (skipped)"

                        echo "Removing agentsmesh/web-admin..."
                        docker rmi agentsmesh/web-admin || echo "Image agentsmesh/web-admin not found (skipped)"

                        echo "Old images removed ✅"
                    """
                }
            }
        }

        stage('Read Configuration') {
            steps {
                script {
                    echo "=== Reading DB_PASSWORD from .env file ==="
                    sh """
                        cd ${DEPLOY_DIR}
                        if ! grep -q "^DB_PASSWORD=" .env; then
                            echo "ERROR: DB_PASSWORD not found in .env file"
                            exit 1
                        fi

                        DB_PASSWORD=\$(grep "^DB_PASSWORD=" .env | cut -d '=' -f2)
                        echo "DB_PASSWORD read successfully ✅"
                        echo "DB_PASSWORD=\${DB_PASSWORD}" > /tmp/deploy_config_${BUILD_NUMBER}.env
                    """
                }
            }
        }

        stage('Start Services') {
            steps {
                script {
                    echo "=== Starting AgentsMesh services ==="
                    sh """
                        cd ${DEPLOY_DIR}
                        echo "Running: docker compose up -d"
                        docker compose up -d

                        echo "Waiting for services to start..."
                        sleep 10

                        echo "Checking service status..."
                        docker compose ps

                        echo "Services started ✅"
                    """
                }
            }
        }

        stage('Wait for Backend Ready') {
            steps {
                script {
                    echo "=== Waiting for backend to be ready ==="
                    sh """
                        cd ${DEPLOY_DIR}

                        TIMEOUT=120
                        ELAPSED=0

                        echo "Waiting for backend health check (timeout: \${TIMEOUT}s)..."

                        while [ \$ELAPSED -lt \$TIMEOUT ]; do
                            if docker compose exec -T backend wget --no-verbose --tries=1 --spider http://localhost:8080/health 2>/dev/null; then
                                echo "Backend is ready ✅"
                                break
                            fi
                            echo "Waiting... (\${ELAPSED}s elapsed)"
                            sleep 5
                            ELAPSED=\$((ELAPSED + 5))
                        done

                        if [ \$ELAPSED -ge \$TIMEOUT ]; then
                            echo "ERROR: Backend health check timed out after \${TIMEOUT}s"
                            echo "Backend logs:"
                            docker compose logs --tail=50 backend
                            exit 1
                        fi
                    """
                }
            }
        }

        stage('Run Database Migration') {
            steps {
                script {
                    echo "=== Running database migrations ==="
                    sh """
                        cd ${DEPLOY_DIR}

                        # Read DB_PASSWORD from temporary config file
                        source /tmp/deploy_config_${BUILD_NUMBER}.env

                        echo "Running migrations..."
                        docker compose exec -T backend migrate -path /app/migrations \\
                            -database "postgres://agentsmesh:\${DB_PASSWORD}@postgres:5432/agentsmesh?sslmode=disable" up

                        echo "Database migrations completed ✅"
                    """
                }
            }
        }
    }

    post {
        success {
            script {
                sh """
                    cd ${DEPLOY_DIR}
                    echo ""
                    echo "=== ✅ Deployment Successful ==="
                    echo ""
                    echo "AgentsMesh services are running:"
                    docker compose ps
                    echo ""
                    echo "To view logs:"
                    echo "  cd ${DEPLOY_DIR}"
                    echo "  docker compose logs -f [service_name]"
                    echo ""
                """
            }
        }
        failure {
            script {
                sh """
                    cd ${DEPLOY_DIR}
                    echo ""
                    echo "=== ❌ Deployment Failed ==="
                    echo ""
                    echo "Recent logs from all services:"
                    docker compose logs --tail=50
                    echo ""
                """
            }
        }
        always {
            script {
                echo "=== Cleaning up temporary files ==="
                sh """
                    rm -f /tmp/deploy_config_${BUILD_NUMBER}.env || true
                """
            }
        }
    }
}
