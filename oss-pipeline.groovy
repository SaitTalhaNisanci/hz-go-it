pipeline {
    agent {
        label "lab"
    }

    parameters {
        string(name: 'NAME', defaultValue: 'runner', description: 'Image name')
    }

    stages {

        stage('Checkout') {
            steps {
                git branch: 'master', credentialsId: '7df9a580-f2f9-4a4a-9523-b22157b6d32f', url: 'https://github.com/hazelcast/hazelcast-go-client/'
                sh "sudo mkdir -p /home/jenkins/go/src/github.com/hazelcast/hazelcast-go-client"
                sh "sudo cp -R ./ /home/jenkins/go/src/github.com/hazelcast/hazelcast-go-client/"
                sh "sudo chmod -R 777 /home/jenkins/go"
            }
        }

        stage('Load Dependencies') {
            steps {
                sh "go get -u golang.org/x/net/context"
                sh "go get -u github.com/docker/libcompose/docker"
                sh "go get -u github.com/docker/libcompose/docker"
                sh "go get -u github.com/docker/docker/api/types"
                sh "go get -u github.com/docker/libcompose/docker/ctx"
                sh "go get -u github.com/docker/libcompose/project"
                sh "go get -u github.com/docker/libcompose/project/options"
                sh "go get -u github.com/stretchr/testify/assert"
                sh "go get -u github.com/lucasjones/reggen"
                sh "go get -u github.com/montanaflynn/stats"
                sh "go get -u github.com/docker/docker/api/types"
                sh "go get -u github.com/docker/docker/api/types/filters"
                sh "go get -u github.com/docker/docker/client"
                sh "ls -ll /home/jenkins/go/src"
            }
        }
        stage('Build Runner') {
            steps {
                git changelog: false, poll: false, url: 'https://github.com/lazerion/hz-go-it.git'
                script {
                    runner = docker.build("${params.NAME}:${env.BUILD_ID}")
                }
            }
        }

        stage('Acceptance') {
            steps {
                sh "docker-compose -f ./acceptance/deployment.yaml down || true"
                sh "docker network create --attachable=true hz-go-it || true"
                sh "docker run --network=hz-go-it --name=hz-go-it -v /home/jenkins/go:/go -v /var/run/docker.sock:/var/run/docker.sock ${params.NAME}:${env.BUILD_ID}"
            }
        }
    }

    post {
        always {
            sh "docker-compose -f ./acceptance/deployment.yaml down || true"
            sh "docker stop hz-go-it || true"
            sh "docker rm hz-go-it || true"
            sh "docker network rm hz-go-it || true"
            sh "sudo rm -rf /home/jenkins/go/*"
            script {
                sh "docker rmi ${runner.id}"
            }
        }
        failure {
            mail to: 'clients@hazelcast.com',
                    subject: "Failed Pipeline: ${currentBuild.fullDisplayName}",
                    body: "Something is wrong with ${env.BUILD_URL}"
        }
    }
}
