@Library('jenkins-shared-libraries') _

pipeline {
    agent {
        kubernetes {
            inheritFrom 'base'
        }
    }
    environment {
        REGISTRY = 'r.glss.ir'
        PROJECT_NAME = 'vod/mashaghel'
        STAGE = 'CI'
        BUILDKIT_PROGRESS = 'plain'
        BUILDX_LOG_LEVEL = 'warn'
        KUSTOMIZE_REPO_URL = 'https://github.com/GolianGroup/deployments.git'
        KUSTOMIZE_DEPLOYMENTS_DIR = 'vod/starnet'
        PROJECT_PATH = '.'
        DOCKERFILE_PATH = "${PROJECT_PATH}/build/dockerfiles/Dockerfile"
        CONTAINER_LABEL = 'go'
        CONTAINER_NAME = 'mashaghel'
        CONTAINER_NEW_NAME = 'r.glss.ir/vod/mashaghel'
    }
    options {
        skipDefaultCheckout(true)
    }
    stages {
        stage('Checkout') {
            steps {
                script {
                    checkoutRepositories(scm, env.KUSTOMIZE_REPO_URL, env.KUSTOMIZE_DEPLOYMENTS_DIR)
                }
            }
        }
        stage('Prepare') {
            steps {
                script {
                    prepareEnvironment()
                }
            }
        }
        stage('buildx') {
            steps {
                script {
                    setupBuildx()
                }
            }
        }
        stage('Build') {
            steps {
                container(env.CONTAINER_LABEL) {
                    script {
                        def tags = []
                        def nextPublicBaseUrl = ""
                        if (env.branchName == 'main') {
                            tags = ["${env.REGISTRY}/${env.PROJECT_NAME}:main-${env.shortCommitHash}", "${env.REGISTRY}/${env.PROJECT_NAME}:main", "${env.REGISTRY}/${env.PROJECT_NAME}:latest"]
                        } else if (env.branchName == 'dev') {
                            tags = ["${env.REGISTRY}/${env.PROJECT_NAME}:staging-${env.shortCommitHash}", "${env.REGISTRY}/${env.PROJECT_NAME}:staging", "${env.REGISTRY}/${env.PROJECT_NAME}:dev-${env.shortCommitHash}", "${env.REGISTRY}/${env.PROJECT_NAME}:dev"]
                        } else {
                            tags = ["${env.REGISTRY}/${env.PROJECT_NAME}:${env.branchName}-${env.shortCommitHash}"]
                        }
                        def buildArgs = []
                        buildAndPushDockerImage(tags, env.DOCKERFILE_PATH, env.PROJECT_PATH, buildArgs)
                        updateKustomizeImageTag(env.branchName, env.shortCommitHash, env.CONTAINER_NAME, env.CONTAINER_NEW_NAME)
                    }
                }
            }
        }
    }
}
