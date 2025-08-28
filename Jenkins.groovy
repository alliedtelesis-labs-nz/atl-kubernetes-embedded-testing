pipeline {
    agent {
        label "${AGENT_LABEL}"
    }

    environment {
        BINARY_NAME = 'ket'
        GO_VERSION = '1.24'
        PROJECT_DIR = 'kubernetes-embedded-testing'
        DOCKER_IMAGE = "${DOCKERHUB_CREDENTIALS_USR}/ket"
    }
    
    options {
         buildDiscarder logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '90', numToKeepStr: '300')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_COMMIT_SHORT = sh(
                        script: 'git rev-parse --short HEAD',
                        returnStdout: true
                    ).trim()
                    env.BUILD_TAG = "${env.BRANCH_NAME}-${env.BUILD_NUMBER}-${env.GIT_COMMIT_SHORT}"
                }
            }
        }
        
        stage('Lint') {
            steps {
                dir("${PROJECT_DIR}") {
                    sh '''
                        golint ./... | tee lint-results.txt
                    '''
                }
            }
        }
        
        stage('Test') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        dir("${PROJECT_DIR}") {
                            sh '''
                                export PATH=/usr/local/go/bin:$PATH
                                go test -v -race -coverprofile=coverage.out ./... | tee test-results.txt
                            '''
                        }
                    }
                }
            }
        }
        
        stage('Build') {
            parallel {
                stage('Build Linux AMD64') {
                    steps {
                        dir("${PROJECT_DIR}") {
                            sh '''
                                export PATH=/usr/local/go/bin:$PATH
                                export GOOS=linux
                                export GOARCH=amd64
                                export CGO_ENABLED=0
                                
                                go build -a -installsuffix cgo -ldflags="-s -w" -o ${BINARY_NAME}-linux-amd64 ./cmd/testrunner
                                tar -czf ${BINARY_NAME}-linux-amd64.tar.gz ${BINARY_NAME}-linux-amd64
                            '''
                        }
                    }
                }
                
                stage('Build Linux ARM64') {
                    steps {
                        dir("${PROJECT_DIR}") {
                            sh '''
                                export PATH=/usr/local/go/bin:$PATH
                                export GOOS=linux
                                export GOARCH=arm64
                                export CGO_ENABLED=0
                                
                                go build -a -installsuffix cgo -ldflags="-s -w" -o ${BINARY_NAME}-linux-arm64 ./cmd/testrunner
                                tar -czf ${BINARY_NAME}-linux-arm64.tar.gz ${BINARY_NAME}-linux-arm64
                            '''
                        }
                    }
                }
                
                stage('Build Darwin AMD64') {
                    steps {
                        dir("${PROJECT_DIR}") {
                            sh '''
                                export PATH=/usr/local/go/bin:$PATH
                                export GOOS=darwin
                                export GOARCH=amd64
                                export CGO_ENABLED=0
                                
                                go build -a -installsuffix cgo -ldflags="-s -w" -o ${BINARY_NAME}-darwin-amd64 ./cmd/testrunner
                                tar -czf ${BINARY_NAME}-darwin-amd64.tar.gz ${BINARY_NAME}-darwin-amd64
                            '''
                        }
                    }
                }
                
                stage('Build Darwin ARM64') {
                    steps {
                        dir("${PROJECT_DIR}") {
                            sh '''
                                export PATH=/usr/local/go/bin:$PATH
                                export GOOS=darwin
                                export GOARCH=arm64
                                export CGO_ENABLED=0
                                
                                go build -a -installsuffix cgo -ldflags="-s -w" -o ${BINARY_NAME}-darwin-arm64 ./cmd/testrunner
                                tar -czf ${BINARY_NAME}-darwin-arm64.tar.gz ${BINARY_NAME}-darwin-arm64
                            '''
                        }
                    }
                }
            }
        }
        
        stage('Publish Docker Image') {
            steps {
                script {
                    dir("${PROJECT_DIR}") {
                        // Build docker image
                        def dockerImageName = "${DOCKER_IMAGE}"
                        dockerImage = docker.build("${dockerImageName}:${env.BUILD_TAG}", "-f Dockerfile .")

                        // Generate version tags
                        def BUILD_DATE = new Date().format('yyyyMMdd-HHmmss')
                        def UNBROKEN_BUILD_DATE = new Date().format('yyyyMMdd')
                        
                        def version = "1.0.0"
                        try {
                            version = sh(
                                script: 'git describe --tags --abbrev=0 2>/dev/null || echo "1.0.0"',
                                returnStdout: true
                            ).trim()
                            // Remove 'v' prefix if present
                            version = version.startsWith('v') ? version.substring(1) : version
                        } catch (Exception e) {
                            echo "No git tags found, using default version: ${version}"
                        }
                        
                        def versions = version.split('\\.')
                        def major = versions[0]
                        def minor = versions.length > 1 ? versions[0] + '.' + versions[1] : major + '.0'
                        def patch = versions.length > 2 ? versions[0] + '.' + versions[1] + '.' + versions[2] : minor + '.0'
                        def full = patch + '_' + BUILD_DATE
                        def staleIdentifier = 'stale-build-ket-' + UNBROKEN_BUILD_DATE

                        def TAGS = [major, minor, patch, full, "ket", staleIdentifier, "latest"]

                         //Publish
                        withDockerRegistry([credentialsId: 'maker-for-docker', url: 'https://docker.atlnz.lc']) {
                            for (TAG in TAGS) {
                                dockerImage.push(TAG)
                            }
                        }
                    }
                }
            }
        }
        
        
        stage('Release') {
            when {
                buildingTag()
            }
            steps {
                script {
                    dir("${PROJECT_DIR}") {
                        sh '''
                            # Install GitHub CLI if not available
                            if ! command -v gh &> /dev/null; then
                                curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
                                echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
                                sudo apt update
                                sudo apt install gh
                            fi
                            
                            # Create release
                            gh release create ${TAG_NAME} *.tar.gz \
                                --title "Release ${TAG_NAME}" \
                                --notes "Automated release created by Jenkins" \
                                --repo ${GIT_URL}
                        '''
                    }
                }
            }
        }
    }
    
    post {
        always {
            cleanWs()
        }
        
        success {
            script {
                if (env.BRANCH_NAME == 'master') {
                    echo "Build successful for branch ${env.BRANCH_NAME}"
                }
            }
        }
        
        failure {
            script {
                echo "Build failed for branch ${env.BRANCH_NAME}"
            }
        }
        
        unstable {
            script {
                echo "Build unstable for branch ${env.BRANCH_NAME}"
            }
        }
    }
}
