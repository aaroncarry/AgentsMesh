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

String getNodeLabel(String env) {
    def envMap = [
        'Test': 'aqa01-i01-jpt44.int.rclabenv.com',
        'Dev': 'aqa01-i01-xta01.int.rclabenv.com'
    ]
    return envMap[env] ?: envMap['Dev']
}

currentBuild.setDescription("triggered by " + getBuildUser())

pipeline {
    agent {
        node {
            label getNodeLabel(params.ENV ?: 'Dev')
        }
    }

    environment {
        // Deployment directory
        DEPLOY_DIR = "${params.DEPLOY_DIR}"
    }

    parameters {
        choice(
            name: 'ENV',
            choices: ['Dev', 'Test'],
            description: 'Target deployment environment'
        )
        string(
            name: 'DEPLOY_DIR',
            defaultValue: '~/workspace/AgentsMesh/deploy/selfhost',
            description: 'AgentsMesh deployment directory'
        )
        string(
            name: 'BACKEND_VERSION',
            defaultValue: '',
            description: 'Backend Docker image version (leave empty to keep current version in .env)'
        )
        string(
            name: 'WEB_VERSION',
            defaultValue: '',
            description: 'Web Docker image version (leave empty to keep current version in .env)'
        )
        string(
            name: 'WEB_ADMIN_VERSION',
            defaultValue: '',
            description: 'Web-Admin Docker image version (leave empty to keep current version in .env)'
        )
        string(
            name: 'RELAY_VERSION',
            defaultValue: '',
            description: 'Relay Docker image version (leave empty to keep current version in .env)'
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
                        echo "Running: docker compose -f docker-compose-rc.yml down"
                        docker compose -f docker-compose-rc.yml down || true
                        echo "Services stopped ✅"
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

        stage('Update Version Configuration') {
            steps {
                script {
                    echo "=== Updating version configuration in .env file ==="
                    sh """
                        cd ${DEPLOY_DIR}

                        # Update BACKEND_VERSION if provided
                        if [ -n "${params.BACKEND_VERSION}" ]; then
                            echo "Updating BACKEND_VERSION to: ${params.BACKEND_VERSION}"
                            if grep -q "^BACKEND_VERSION=" .env; then
                                sed -i.bak "s|^BACKEND_VERSION=.*|BACKEND_VERSION=${params.BACKEND_VERSION}|" .env
                            else
                                echo "BACKEND_VERSION=${params.BACKEND_VERSION}" >> .env
                            fi
                        else
                            echo "BACKEND_VERSION not provided, keeping current value"
                        fi

                        # Update WEB_VERSION if provided
                        if [ -n "${params.WEB_VERSION}" ]; then
                            echo "Updating WEB_VERSION to: ${params.WEB_VERSION}"
                            if grep -q "^WEB_VERSION=" .env; then
                                sed -i.bak "s|^WEB_VERSION=.*|WEB_VERSION=${params.WEB_VERSION}|" .env
                            else
                                echo "WEB_VERSION=${params.WEB_VERSION}" >> .env
                            fi
                        else
                            echo "WEB_VERSION not provided, keeping current value"
                        fi

                        # Update WEB_ADMIN_VERSION if provided
                        if [ -n "${params.WEB_ADMIN_VERSION}" ]; then
                            echo "Updating WEB_ADMIN_VERSION to: ${params.WEB_ADMIN_VERSION}"
                            if grep -q "^WEB_ADMIN_VERSION=" .env; then
                                sed -i.bak "s|^WEB_ADMIN_VERSION=.*|WEB_ADMIN_VERSION=${params.WEB_ADMIN_VERSION}|" .env
                            else
                                echo "WEB_ADMIN_VERSION=${params.WEB_ADMIN_VERSION}" >> .env
                            fi
                        else
                            echo "WEB_ADMIN_VERSION not provided, keeping current value"
                        fi

                        # Update RELAY_VERSION if provided
                        if [ -n "${params.RELAY_VERSION}" ]; then
                            echo "Updating RELAY_VERSION to: ${params.RELAY_VERSION}"
                            if grep -q "^RELAY_VERSION=" .env; then
                                sed -i.bak "s|^RELAY_VERSION=.*|RELAY_VERSION=${params.RELAY_VERSION}|" .env
                            else
                                echo "RELAY_VERSION=${params.RELAY_VERSION}" >> .env
                            fi
                        else
                            echo "RELAY_VERSION not provided, keeping current value"
                        fi

                        echo ""
                        echo "Current version configuration:"
                        grep -E "^(BACKEND_VERSION|WEB_VERSION|WEB_ADMIN_VERSION|RELAY_VERSION)=" .env || echo "No version variables found in .env"
                        echo ""
                        echo "Version configuration updated ✅"
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
                        echo "Running: docker compose -f docker-compose-rc.yml up -d"
                        docker compose -f docker-compose-rc.yml up -d

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
